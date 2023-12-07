package gamevrf

// Specifies a domain for randomness generation.
type DomainSeparationTag int64

const (
	DomainSeparationTag_GameBasic DomainSeparationTag = 1 + iota
	// dedicate for a round
	DomainSeparationTag_GameRound
	// dedicate for lottery
	DomainSeparationTag_GameLottery
	// dedicate for players
	DomainSeparationTag_GamePlayers
)
