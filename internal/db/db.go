package db

type DB interface {
	GuildSettings() GuildSettings
	LikedSongs() LikedSongs
	MusicPlayers() MusicPlayers
	PlayHistory() PlayHistory
	Tags() Tags
	Voters() Voters
}
