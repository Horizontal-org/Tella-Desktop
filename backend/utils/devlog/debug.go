//go:build !production

package devlog
func IsDevelop() bool {
	return true
}
