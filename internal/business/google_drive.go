package business

import (
	"context"
	"errors"
	"fmt"
	"os"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
	"gorm.io/gorm"
)

func getGoogleOAuthConfig() *oauth2.Config {
	return &oauth2.Config{
		ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		RedirectURL:  os.Getenv("GOOGLE_REDIRECT_URL"),
		Scopes: []string{
			"https://www.googleapis.com/auth/drive.file",
		},
		Endpoint: google.Endpoint,
	}
}

// GetGoogleLoginURL generates the URL for the user to login with Google
func GetGoogleLoginURL(businessID uint) string {
	config := getGoogleOAuthConfig()
	// Pass businessID in state to identify which business is connecting
	return config.AuthCodeURL(fmt.Sprintf("%d", businessID), oauth2.AccessTypeOffline, oauth2.ApprovalForce)
}

// HandleGoogleCallback handles the code returned from Google OAuth
func HandleGoogleCallback(db *gorm.DB, businessID uint, code string) error {
	config := getGoogleOAuthConfig()
	token, err := config.Exchange(context.Background(), code)
	if err != nil {
		return err
	}

	var biz Business
	if err := db.First(&biz, businessID).Error; err != nil {
		return err
	}

	biz.GoogleAccessToken = token.AccessToken
	biz.GoogleRefreshToken = token.RefreshToken
	biz.GoogleTokenExpiry = &token.Expiry
	biz.GoogleDriveLinked = true

	// Step: Create the backup folder in their Google Drive
	folderID, err := createBackupFolder(token, config, biz.Name)
	if err != nil {
		return fmt.Errorf("failed to create backup folder: %w", err)
	}
	biz.GoogleDriveFolderID = folderID

	return db.Save(&biz).Error
}

func createBackupFolder(token *oauth2.Token, config *oauth2.Config, businessName string) (string, error) {
	ctx := context.Background()
	client := config.Client(ctx, token)
	srv, err := drive.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return "", err
	}

	folderName := fmt.Sprintf("AlphaKit_Backups_%s", businessName)

	// Check if folder already exists (optional, but good practice)
	// For now, let's just create a new one
	f := &drive.File{
		Name:     folderName,
		MimeType: "application/vnd.google-apps.folder",
	}

	res, err := srv.Files.Create(f).Fields("id").Do()
	if err != nil {
		return "", err
	}

	return res.Id, nil
}

// RefreshGoogleToken ensures the token is valid, refreshing it if necessary
func RefreshGoogleToken(db *gorm.DB, biz *Business) (*oauth2.Token, error) {
	if biz.GoogleRefreshToken == "" {
		return nil, errors.New("no refresh token available")
	}

	config := getGoogleOAuthConfig()
	token := &oauth2.Token{
		AccessToken:  biz.GoogleAccessToken,
		RefreshToken: biz.GoogleRefreshToken,
		Expiry:       *biz.GoogleTokenExpiry,
	}

	tokenSource := config.TokenSource(context.Background(), token)
	newToken, err := tokenSource.Token()
	if err != nil {
		return nil, err
	}

	// Update DB if token was actually refreshed
	if newToken.AccessToken != biz.GoogleAccessToken {
		biz.GoogleAccessToken = newToken.AccessToken
		biz.GoogleTokenExpiry = &newToken.Expiry
		db.Model(biz).Updates(map[string]interface{}{
			"google_access_token": biz.GoogleAccessToken,
			"google_token_expiry": biz.GoogleTokenExpiry,
		})
	}

	return newToken, nil
}

// GetTokenSource returns an oauth2.TokenSource for the given token
func GetTokenSource(token *oauth2.Token) oauth2.TokenSource {
	config := getGoogleOAuthConfig()
	return config.TokenSource(context.Background(), token)
}
