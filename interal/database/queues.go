package database

import (
	"github.com/disgoorg/disgolink/v2/lavalink"
	"github.com/disgoorg/json"
	"github.com/disgoorg/snowflake/v2"
)

type Track struct {
	ID         int            `db:"id"`
	GuildID    snowflake.ID   `db:"guild_id"`
	Position   int            `db:"position"`
	Encoded    string         `db:"encoded"`
	Info       []byte         `db:"info"`
	PluginInfo []byte         `db:"plugin_info"`
	Track      lavalink.Track `db:"-"`
}

func (t *Track) Marshal() error {
	infoBuf, err := json.Marshal(t.Track.Info)
	if err != nil {
		return err
	}
	pluginInfoBuf, err := json.Marshal(t.Track.PluginInfo)
	if err != nil {
		return err
	}
	t.Info = infoBuf
	t.PluginInfo = pluginInfoBuf
	return nil
}

func (t *Track) Unmarshal() error {
	t.Track.Encoded = t.Encoded
	err := json.Unmarshal(t.Info, &t.Track.Info)
	if err != nil {
		return err
	}
	err = json.Unmarshal(t.PluginInfo, &t.Track.PluginInfo)
	if err != nil {
		return err
	}
	return nil
}

func (d *Database) GetQueue(guildID snowflake.ID) ([]Track, error) {
	var queue []Track
	if err := d.dbx.Select(&queue, "SELECT * FROM queues WHERE guild_id = $1 ORDER BY position DESC", guildID); err != nil {
		return nil, err
	}

	for i := range queue {
		if err := queue[i].Unmarshal(); err != nil {
			return nil, err
		}
	}
	return queue, nil
}

func (d *Database) AddTracks(guildID snowflake.ID, tracks []lavalink.Track) error {
	dbTracks := make([]Track, len(tracks))
	for i, track := range tracks {
		dbTrack := Track{
			GuildID: guildID,
			Encoded: track.Encoded,
			Track:   track,
		}

		if err := dbTrack.Marshal(); err != nil {
			return err
		}
		dbTracks[i] = dbTrack
	}

	_, err := d.dbx.NamedExec("INSERT INTO queues (guild_id, encoded, info, plugin_info) VALUES (:guild_id, :encoded, :info, :plugin_info)", dbTracks)
	return err
}

func (d *Database) MoveTrack(position int, newPosition int) error {
	_, err := d.dbx.Exec("UPDATE queues SET position = $1 WHERE id = $2", position, newPosition)
	return err
}

func (d *Database) NextTrack(guildID snowflake.ID) (*Track, error) {
	var track Track
	err := d.dbx.Get(&track, "DELETE FROM queues WHERE position = (SELECT MIN(position) from queues WHERE guild_id = $1) RETURNING *", guildID)
	if err != nil {
		return nil, err
	}
	if err = track.Unmarshal(); err != nil {
		return nil, err
	}

	return &track, nil
}

func (d *Database) RemoveTrack(trackID int) error {
	_, err := d.dbx.Exec("DELETE FROM queues WHERE id = $1", trackID)
	return err
}

func (d *Database) ClearQueue(guildID snowflake.ID) error {
	_, err := d.dbx.Exec("DELETE FROM queues WHERE guild_id = $1", guildID)
	return err
}

func (d *Database) ShuffleQueue(guildID snowflake.ID) error {
	var queueSize int
	err := d.dbx.Get(&queueSize, "SELECT COUNT(*) FROM queues WHERE guild_id = $1", guildID)
	if err != nil {
		return err
	}
	_, err = d.dbx.Exec("UPDATE queues SET position = floor(random() * $1) + 1 WHERE guild_id = $2", queueSize, guildID)
	return err
}
