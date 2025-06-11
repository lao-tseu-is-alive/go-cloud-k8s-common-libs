package goHttpEcho

type AppInfo struct {
	App        string `json:"app"`
	Version    string `json:"version"`
	Repository string `json:"repository"`
	AuthUrl    string `json:"authUrl"`
}

type VersionReader interface {
	GetAppInfo() AppInfo
}

// SimpleVersionWriter Create a struct that will implement the VersionReader interface
type SimpleVersionWriter struct {
	Info AppInfo
}

// GetAppInfo returns the app information of the application.
func (s SimpleVersionWriter) GetAppInfo() AppInfo {
	return s.Info
}

// NewSimpleVersionReader is a constructor that initializes the VersionReader interface
func NewSimpleVersionReader(app, ver, repo, authUrl string) *SimpleVersionWriter {

	return &SimpleVersionWriter{
		Info: AppInfo{
			App:        app,
			Version:    ver,
			Repository: repo,
			AuthUrl:    authUrl,
		},
	}
}
