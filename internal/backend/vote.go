package backend

import (
	"context"
	"time"

	"github.com/disgoorg/snowflake"
)

func (b *Backend) AddVote(userID snowflake.Snowflake, botList BotList, multiplier int) error {
	voteDuration := botList.VoteCooldown * 2 * time.Duration(multiplier)
	if err := b.DB.Voters().Add(userID, voteDuration); err != nil {
		return err
	}
	return b.Rest.AddMemberRole(b.Config.SupportGuildID, userID, b.Config.BotLists.VoterRoleID)
}

func (b *Backend) VoteTask(ctx context.Context) {
	voters, err := b.DB.Voters().GetAll(time.Now())
	if err != nil {
		b.Logger.Error("failed to fetch expired votes: ", err)
		return
	}
	for _, voter := range voters {
		if err = b.Rest.RemoveMemberRole(b.Config.SupportGuildID, snowflake.Snowflake(voter.UserID), b.Config.BotLists.VoterRoleID); err != nil {
			b.Logger.Error("failed to remove voter role: ", err)
		}
	}
}
