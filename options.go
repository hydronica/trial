package trial

type Options interface {
	set(o *ComparerOption) bool
}
