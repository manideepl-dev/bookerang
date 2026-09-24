package domain

type AddBookResult struct {
	CopyID string
	Added  bool
}

type Copy struct {
	CopyID string
	Title  string
	Author string
}

type NearbyBook struct {
	CopyID         string
	Title          string
	Author         string
	OwnerUsername  string
	OwnerFirstName string
	OwnerLastName  string
}
