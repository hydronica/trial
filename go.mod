module github.com/hydronica/trial

go 1.18

// Note: go-cmp v0.7.0 is available but requires Go 1.21+.
// We use v0.6.0 to support Go 1.18+ (minimum for generics).
// v0.7.0 only adds compare function support for SortSlices/SortMaps,
// which is not needed by this library.
require github.com/google/go-cmp v0.6.0
