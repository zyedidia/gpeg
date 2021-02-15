package memo

// An Edit represents a modification to the subject string where the interval
// [Start, End) is modified to be Len bytes. If Len = 0, this is equivalent
// to deleting the interval, and if Start = End this is an insertion.
type Edit struct {
	Start, End int
	Len        int
}
