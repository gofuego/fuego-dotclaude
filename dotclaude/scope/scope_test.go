package scope

import (
	"path/filepath"
	"testing"
)

// fakeFS treats the listed paths (cleaned) as existing directories.
type fakeFS map[string]bool

func (f fakeFS) IsDir(p string) bool { return f[filepath.Clean(p)] }

func ptr(b bool) *bool { return &b }

func TestResolve(t *testing.T) {
	tests := []struct {
		name      string
		fs        fakeFS
		arg       string
		home      string
		flag      *bool
		wantDir   string
		wantSib   bool
		wantMode  string
		wantError bool
	}{
		{
			name:     "no arg defaults to home/.claude isolated",
			fs:       fakeFS{},
			arg:      "",
			home:     "/home/u",
			wantDir:  "/home/u/.claude",
			wantSib:  false,
			wantMode: "isolated",
		},
		{
			name:     "arg is a .claude dir is isolated",
			fs:       fakeFS{"/x/.claude": true},
			arg:      "/x/.claude",
			wantDir:  "/x/.claude",
			wantSib:  false,
			wantMode: "isolated",
		},
		{
			name:     "arg containing .claude is a project with siblings",
			fs:       fakeFS{"/repo": true, "/repo/.claude": true},
			arg:      "/repo",
			wantDir:  "/repo/.claude",
			wantSib:  true,
			wantMode: "project",
		},
		{
			name:     "project with --no-siblings",
			fs:       fakeFS{"/repo": true, "/repo/.claude": true},
			arg:      "/repo",
			flag:     ptr(false),
			wantDir:  "/repo/.claude",
			wantSib:  false,
			wantMode: "project",
		},
		{
			name:     "isolated .claude with --siblings forced on",
			fs:       fakeFS{"/x/.claude": true},
			arg:      "/x/.claude",
			flag:     ptr(true),
			wantDir:  "/x/.claude",
			wantSib:  true,
			wantMode: "isolated",
		},
		{
			name:     "plain dir treated as content, isolated",
			fs:       fakeFS{"/custom": true},
			arg:      "/custom",
			wantDir:  "/custom",
			wantSib:  false,
			wantMode: "isolated",
		},
		{
			name:      "missing path errors",
			fs:        fakeFS{},
			arg:       "/missing",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, err := Resolve(tt.fs, tt.arg, tt.home, tt.flag)
			if tt.wantError {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if r.ContentDir != tt.wantDir {
				t.Errorf("ContentDir = %q, want %q", r.ContentDir, tt.wantDir)
			}
			if r.Siblings != tt.wantSib {
				t.Errorf("Siblings = %v, want %v", r.Siblings, tt.wantSib)
			}
			if r.Mode != tt.wantMode {
				t.Errorf("Mode = %q, want %q", r.Mode, tt.wantMode)
			}
		})
	}
}
