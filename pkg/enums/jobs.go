package enums

type JobType string

const JobTypeList JobType = "list"

type ListKind string

const (
	ListKindPublicList ListKind = "public_list"
	ListKindWatchlist  ListKind = "watchlist"
)

func (kind ListKind) Valid() bool {
	return kind == ListKindPublicList || kind == ListKindWatchlist
}

type ListOrder string

const (
	ListOrderSource ListOrder = "source"
	ListOrderRank   ListOrder = "rank"
)

func (order ListOrder) Valid() bool {
	return order == "" || order == ListOrderSource || order == ListOrderRank
}
