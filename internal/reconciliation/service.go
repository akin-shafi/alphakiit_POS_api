package reconciliation

import (
	"fmt"
	"pos-fiber-app/internal/sale"
	"strings"
	"time"

	"gorm.io/gorm"
)

type ReconciliationService struct {
	db        *gorm.DB
	providers map[string]PaymentProvider
}

func NewReconciliationService(db *gorm.DB) *ReconciliationService {
	return &ReconciliationService{
		db:        db,
		providers: make(map[string]PaymentProvider),
	}
}

func (s *ReconciliationService) RegisterProvider(p PaymentProvider) {
	s.providers[strings.ToLower(p.GetName())] = p
}

func (s *ReconciliationService) HandleWebhook(providerName string, payload []byte, headers map[string]string) error {
	provider, ok := s.providers[strings.ToLower(providerName)]
	if !ok {
		return ErrProviderNotFound
	}

	normalized, err := provider.VerifyWebhook(payload, headers)
	if err != nil {
		return err
	}

	var payment sale.Payment
	if err := s.db.Where("internal_reference = ?", normalized.InternalRef).First(&payment).Error; err != nil {
		s.db.Create(&sale.PaymentLog{
			BusinessID: 0,
			Provider:   providerName,
			RawPayload: string(payload),
			Status:     "UNLINKED",
			Notes:      fmt.Sprintf("No payment found for ref: %s", normalized.InternalRef),
		})
		return fmt.Errorf("payment not found for reference: %s", normalized.InternalRef)
	}

	payment.ExternalReference = normalized.ExternalRef
	payment.Metadata = normalized.Raw
	payment.RawResponse = string(payload)
	payment.ReconciledAt = time.Now()
	payment.HardwareTerminalID = normalized.HardwareID
	payment.CommissionFee = normalized.Fee
	payment.NetAmount = normalized.Amount - normalized.Fee
	
	if normalized.Status != "SUCCESS" {
		payment.Status = sale.ReconFailed
		s.db.Save(&payment)
		return nil
	}

	if normalized.Amount < payment.Amount {
		payment.Status = sale.ReconPartial
	} else if normalized.Amount > payment.Amount {
		payment.Status = sale.ReconMismatch
	} else {
		payment.Status = sale.ReconSuccess
	}

	if err := s.db.Save(&payment).Error; err != nil {
		return err
	}

	if payment.Status == sale.ReconSuccess {
		err := s.db.Model(&sale.Sale{}).Where("id = ?", payment.SaleID).Update("status", sale.StatusCompleted).Error
		if err == nil {
			sale.GlobalKDSHub.BroadcastOrder(payment.BusinessID, sale.EventPaymentVerified, map[string]interface{}{
				"internal_reference": payment.InternalReference,
				"sale_id":           payment.SaleID,
				"status":            payment.Status,
			})
		}
	}

	return nil
}

func (s *ReconciliationService) GetPaymentStatus(reference string) (string, error) {
	var p sale.Payment
	if err := s.db.Where("internal_reference = ?", reference).First(&p).Error; err != nil {
		return "", err
	}
	return string(p.Status), nil
}

func (s *ReconciliationService) GetPayments(businessID uint) ([]sale.Payment, error) {
	var payments []sale.Payment
	err := s.db.Where("business_id = ?", businessID).Order("created_at desc").Limit(100).Find(&payments).Error
	return payments, err
}

func (s *ReconciliationService) GetLogs(businessID uint) ([]sale.PaymentLog, error) {
	var logs []sale.PaymentLog
	err := s.db.Where("business_id = ?", businessID).Order("created_at desc").Limit(100).Find(&logs).Error
	return logs, err
}

type ReconciliationSummary struct {
	TotalCount      int            `json:"total_count"`
	TotalAmount     float64        `json:"total_amount"`
	TotalCommission float64        `json:"total_commission"`
	TotalNet        float64        `json:"total_net"`
	StatusCounts    map[string]int `json:"status_counts"`
	ProviderStats   map[string]int `json:"provider_stats"`
	RecentAlerts    int            `json:"recent_alerts"`
}

func (s *ReconciliationService) GetSummary(businessID uint) (*ReconciliationSummary, error) {
	var payments []sale.Payment
	thirtyDaysAgo := time.Now().AddDate(0, 0, -30)
	err := s.db.Where("business_id = ? AND created_at > ?", businessID, thirtyDaysAgo).Find(&payments).Error
	if err != nil {
		return nil, err
	}

	summary := &ReconciliationSummary{
		StatusCounts:  make(map[string]int),
		ProviderStats: make(map[string]int),
	}

	for _, p := range payments {
		summary.TotalCount++
		summary.TotalAmount += p.Amount
		summary.TotalCommission += p.CommissionFee
		summary.TotalNet += p.NetAmount
		summary.StatusCounts[string(p.Status)]++
		summary.ProviderStats[p.Provider]++
		
		if (p.Status == sale.ReconMismatch || p.Status == sale.ReconPartial) && p.CreatedAt.After(time.Now().Add(-24*time.Hour)) {
			summary.RecentAlerts++
		}
	}

	return summary, nil
}

func (s *ReconciliationService) ManuallyVerify(paymentID uint, reason string) error {
	var p sale.Payment
	if err := s.db.First(&p, paymentID).Error; err != nil {
		return err
	}

	p.Status = sale.ReconSuccess
	p.ReconciledAt = time.Now()
	p.RawResponse = fmt.Sprintf("MANUAL_OVERRIDE: %s", reason)

	if err := s.db.Save(&p).Error; err != nil {
		return err
	}

	err := s.db.Model(&sale.Sale{}).Where("id = ?", p.SaleID).Update("status", sale.StatusCompleted).Error
	if err == nil {
		sale.GlobalKDSHub.BroadcastOrder(p.BusinessID, sale.EventPaymentVerified, map[string]interface{}{
			"internal_reference": p.InternalReference,
			"sale_id":           p.SaleID,
			"status":            p.Status,
		})
	}
	return err
}

type DailySettlement struct {
	Date           time.Time          `json:"date"`
	TotalVolume    float64            `json:"total_volume"`
	TotalFees      float64            `json:"total_fees"`
	NetSettlement  float64            `json:"net_settlement"`
	TerminalCounts map[string]int     `json:"terminal_counts"`
	ProviderVolume map[string]float64 `json:"provider_volume"`
}

func (s *ReconciliationService) GenerateDailySettlement(businessID uint, date time.Time) (*DailySettlement, error) {
	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	var payments []sale.Payment
	err := s.db.Where("business_id = ? AND created_at >= ? AND created_at < ? AND status = ?", businessID, startOfDay, endOfDay, sale.ReconSuccess).Find(&payments).Error
	if err != nil {
		return nil, err
	}

	settlement := &DailySettlement{
		Date:           startOfDay,
		TerminalCounts: make(map[string]int),
		ProviderVolume: make(map[string]float64),
	}

	for _, p := range payments {
		settlement.TotalVolume += p.Amount
		settlement.TotalFees += p.CommissionFee
		settlement.NetSettlement += p.NetAmount
		
		if p.HardwareTerminalID != "" {
			settlement.TerminalCounts[p.HardwareTerminalID]++
		}
		settlement.ProviderVolume[p.Provider] += p.Amount
	}

	return settlement, nil
}
