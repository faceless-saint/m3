package mod

// InitError indicates a problem turning an imported Raw mod into a full
// mod implementation, or an invalid mod property assignment.
type InitError struct {
	Mod     Raw
	message string
}

func (this *InitError) Error() string { return this.message }
