package domain

func NewDebateSide(
	iD int,
	debateID int,
	name string,
	description *string,
	displayOrder int,
) DebateSide {
	return DebateSide{
		iD,
		debateID,
		name,
		description,
		displayOrder,
	}
}

type DebateSide struct {
	ID           int
	DebateID     int
	Name         string
	Description  *string
	DisplayOrder int
}
