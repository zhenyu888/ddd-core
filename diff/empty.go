package diff

func EmptyListDiff() ListDiff {
	return &emptyListDiff{}
}

type emptyListDiff struct{}

func (e emptyListDiff) IsChanged() bool {
	return false
}

func (e emptyListDiff) Added() []interface{} {
	return nil
}

func (e emptyListDiff) Removed() []interface{} {
	return nil
}

func (e emptyListDiff) Modified() []interface{} {
	return nil
}

func EmptyAggregateDiff() AggregateDiff {
	return &emptyAggregateDiff{}
}

type emptyAggregateDiff struct{}

func (e emptyAggregateDiff) IsEmpty() bool {
	return true
}

func (e emptyAggregateDiff) IsSelfChanged() bool {
	return false
}

func (e emptyAggregateDiff) GetDiff(s string) Diff {
	return NewDiff(false)
}

func (e emptyAggregateDiff) GetListDiff(s string) ListDiff {
	return EmptyListDiff()
}
