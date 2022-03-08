package types

import (
	"context"
	"time"

	"github.com/DisgoOrg/snowflake"
	"github.com/KittyBot-Org/KittyBotGo/internal/models"
)

func (b *Backend) AddVote(userID snowflake.Snowflake, botList BotList, multiplier int) error {
	voteDuration := botList.VoteCooldown * 2 * time.Duration(multiplier)
	voter := models.Voter{
		ID:        userID,
		ExpiresAt: time.Now().Add(voteDuration),
	}
	if _, err := b.DB.NewInsert().Model(&voter).
		On("CONFLICT user_id DO UPDATE").
		Set("expires_at = expires_at + ?", voteDuration).
		Exec(context.TODO()); err != nil {
		return err
	}
	return b.RestServices.GuildService().AddMemberRole(b.Config.SupportGuildID, userID, b.Config.BotLists.VoterRoleID)
}

func (b *Backend) VoteTask(ctx context.Context) {
	var voters []models.Voter
	_, err := b.DB.NewSelect().Model(&voters).Where("expires_at < ?", time.Now()).Exec(ctx)
	if err != nil {
		b.Logger.Error("failed to fetch expired votes: ", err)
		return
	}
	for _, voter := range voters {
		if err = b.RestServices.GuildService().RemoveMemberRole(b.Config.SupportGuildID, voter.ID, b.Config.BotLists.VoterRoleID); err != nil {
			b.Logger.Error("failed to remove voter role: ", err)
		}
	}
}
