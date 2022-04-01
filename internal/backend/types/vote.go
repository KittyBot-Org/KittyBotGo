package types

import (
	"context"
	"time"

	"github.com/KittyBot-Org/KittyBotGo/internal/db"
	"github.com/disgoorg/snowflake"
)

func (b *Backend) AddVote(userID snowflake.Snowflake, botList BotList, multiplier int) error {
	voteDuration := botList.VoteCooldown * 2 * time.Duration(multiplier)
	voter := db.Voter{
		ID:        userID,
		ExpiresAt: time.Now().Add(voteDuration),
	}
	if _, err := b.DB.NewInsert().Model(&voter).
		On("CONFLICT user_id DO UPDATE").
		Set("expires_at = expires_at + ?", voteDuration).
		Exec(context.TODO()); err != nil {
		return err
	}
	return b.Rest.Members().AddMemberRole(b.Config.SupportGuildID, userID, b.Config.BotLists.VoterRoleID)
}

func (b *Backend) VoteTask(ctx context.Context) {
	var voters []db.Voter
	_, err := b.DB.NewSelect().Model(&voters).Where("expires_at < ?", time.Now()).Exec(ctx)
	if err != nil {
		b.Logger.Error("failed to fetch expired votes: ", err)
		return
	}
	for _, voter := range voters {
		if err = b.Rest.Members().RemoveMemberRole(b.Config.SupportGuildID, voter.ID, b.Config.BotLists.VoterRoleID); err != nil {
			b.Logger.Error("failed to remove voter role: ", err)
		}
	}
}