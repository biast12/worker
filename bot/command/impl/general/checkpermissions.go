package general

import (
	"fmt"
	"time"

	"github.com/TicketsBot-cloud/common/permission"
	"github.com/TicketsBot-cloud/gdl/objects/channel/embed"
	"github.com/TicketsBot-cloud/gdl/objects/interaction"
	permissiongdl "github.com/TicketsBot-cloud/gdl/permission"
	"github.com/TicketsBot-cloud/worker/bot/command"
	"github.com/TicketsBot-cloud/worker/bot/command/registry"
	"github.com/TicketsBot-cloud/worker/bot/customisation"
	"github.com/TicketsBot-cloud/worker/bot/dbclient"
	"github.com/TicketsBot-cloud/worker/bot/permissionwrapper"
	"github.com/TicketsBot-cloud/worker/i18n"
)

type CheckPermissionsCommand struct {
}

func (CheckPermissionsCommand) Properties() registry.Properties {
	return registry.Properties{
		Name:             "checkpermissions",
		Description:      i18n.HelpCheckPermissions,
		Type:             interaction.ApplicationCommandTypeChatInput,
		PermissionLevel:  permission.Support,
		Category:         command.General,
		DefaultEphemeral: true,
		Timeout:          time.Second * 10,
	}
}

func (c CheckPermissionsCommand) GetExecutor() interface{} {
	return c.Execute
}

type PermissionCheck struct {
	Permission  permissiongdl.Permission
	Description i18n.MessageId
}

type permResult struct {
	Name       string
	Id         uint64
	Missing    []PermissionCheck
	IsCategory bool
}

var (
	// All permissions that are checked for the bot
	allPermissions = []PermissionCheck{
		{permissiongdl.ViewChannel, i18n.PermissionsReadMessages},
		{permissiongdl.ManageChannels, i18n.PermissionsManageChannels},
		{permissiongdl.ManageRoles, i18n.PermissionsManageRoles},
		{permissiongdl.ManageWebhooks, i18n.PermissionsManageWebhooks},
		{permissiongdl.SendMessages, i18n.PermissionsSendMessages},
		{permissiongdl.SendMessagesInThreads, i18n.PermissionsSendMessagesInThreads},
		{permissiongdl.CreatePublicThreads, i18n.PermissionsCreatePublicThreads},
		{permissiongdl.CreatePrivateThreads, i18n.PermissionsCreatePrivateThreads},
		{permissiongdl.EmbedLinks, i18n.PermissionsEmbedLinks},
		{permissiongdl.AttachFiles, i18n.PermissionsAttachFiles},
		{permissiongdl.AddReactions, i18n.PermissionsAddReactions},
		{permissiongdl.UseExternalEmojis, i18n.PermissionsUseExternalEmojis},
		{permissiongdl.MentionEveryone, i18n.PermissionsMentionEveryone},
		{permissiongdl.ManageThreads, i18n.PermissionsManageThreads},
		{permissiongdl.ReadMessageHistory, i18n.PermissionsReadMessageHistory},
		{permissiongdl.UseApplicationCommands, i18n.PermissionsUseApplicationCommands},
	}
	// For categories, only need ManageChannels and ViewChannel
	categoryPermissions = []PermissionCheck{
		{permissiongdl.ViewChannel, i18n.PermissionsReadMessages},
		{permissiongdl.ManageChannels, i18n.PermissionsManageChannels},
	}
	// For transcript channels, need ViewChannel, SendMessages, EmbedLinks, AttachFiles
	transcriptChannelPermissions = []PermissionCheck{
		{permissiongdl.ViewChannel, i18n.PermissionsReadMessages},
		{permissiongdl.SendMessages, i18n.PermissionsSendMessages},
		{permissiongdl.EmbedLinks, i18n.PermissionsEmbedLinks},
		{permissiongdl.AttachFiles, i18n.PermissionsAttachFiles},
	}
	// For notification channels, need ViewChannel, SendMessages
	notificationChannelPermissions = []PermissionCheck{
		{permissiongdl.ViewChannel, i18n.PermissionsReadMessages},
		{permissiongdl.SendMessages, i18n.PermissionsSendMessages},
	}
)

func (CheckPermissionsCommand) Execute(ctx registry.CommandContext) {
	worker := ctx.Worker()
	guildId := ctx.GuildId()
	botId := worker.BotId

	// Check for Administrator permission
	if permissionwrapper.HasPermissions(worker, guildId, botId, permissiongdl.Administrator) {
		ctx.Reply(customisation.Green, i18n.PermissionsTitle, i18n.PermissionsHasAdministrator, botId)
		return
	}

	var results []permResult

	settings, err := ctx.Settings()
	if err != nil {
		ctx.HandleError(err)
		return
	}

	// Track checked channel/category IDs to avoid duplicate checks
	checkedIds := make(map[uint64]struct{})

	// Helper to check a channel/category with a custom permission set
	checkPerms := func(id uint64, isCategory bool, name string, perms []PermissionCheck) {
		if _, checked := checkedIds[id]; checked {
			return
		}
		checkedIds[id] = struct{}{}
		missing := []PermissionCheck{}
		for _, check := range perms {
			if !permissionwrapper.HasPermissionsChannel(worker, guildId, botId, id, check.Permission) {
				missing = append(missing, check)
			}
		}
		if len(missing) > 0 {
			results = append(results, permResult{Name: name, Id: id, Missing: missing, IsCategory: isCategory})
		}
	}

	// Check ticket category
	if categoryId, err := dbclient.Client.ChannelCategory.Get(ctx, guildId); err == nil && categoryId != 0 {
		ch, err := worker.GetChannel(categoryId)
		if err != nil {
			ctx.HandleError(err)
			return
		}
		checkPerms(categoryId, true, ch.Name, categoryPermissions)
	}

	// Check transcript channel
	if archiveId, err := dbclient.Client.ArchiveChannel.Get(ctx, guildId); err == nil && archiveId != nil {
		channel, err := worker.GetChannel(*archiveId)
		if err != nil {
			ctx.HandleError(err)
			return
		}
		checkPerms(*archiveId, false, channel.Name, transcriptChannelPermissions)
	}

	// Check notification channel
	if settings.UseThreads && settings.TicketNotificationChannel != nil {
		channel, err := worker.GetChannel(*settings.TicketNotificationChannel)
		if err != nil {
			ctx.HandleError(err)
			return
		}
		checkPerms(*settings.TicketNotificationChannel, false, channel.Name, notificationChannelPermissions)
	}

	// Check overflow category
	if settings.OverflowEnabled && settings.OverflowCategoryId != nil {
		channel, err := worker.GetChannel(*settings.OverflowCategoryId)
		if err != nil {
			ctx.HandleError(err)
			return
		}
		checkPerms(*settings.OverflowCategoryId, true, channel.Name, categoryPermissions)
	}

	// Check bot permissions (guild-wide)
	missingBotPermissions := []PermissionCheck{}
	for _, check := range allPermissions {
		if !permissionwrapper.HasPermissions(worker, guildId, botId, check.Permission) {
			missingBotPermissions = append(missingBotPermissions, check)
		}
	}

	// Build response
	if len(missingBotPermissions) == 0 && len(results) == 0 {
		ctx.Reply(customisation.Green, i18n.PermissionsTitle, i18n.PermissionsHasAll, botId)
	} else {
		embed := embed.NewEmbed().
			SetTitle(ctx.GetMessage(i18n.PermissionsMissing)).
			SetColor(ctx.GetColour(customisation.Red))

		separator := "────────────"
		blank := "ㅤ"

		sectionCount := 0
		totalSections := 0
		if len(missingBotPermissions) > 0 {
			totalSections++
		}
		totalSections += len(results)

		maxFields := 25
		fieldCounter := 0
		more := false

		if len(missingBotPermissions) > 0 && fieldCounter < maxFields {
			fieldName := fmt.Sprintf("Server-wide\n%s", separator)
			fieldValue := ""
			for _, missing := range missingBotPermissions {
				fieldValue += fmt.Sprintf("**%s**\n%s\n", missing.Permission.String(), ctx.GetMessage(missing.Description))
			}
			sectionCount++
			// Add blank line at end unless last section
			if sectionCount < totalSections {
				fieldValue += blank
			}
			embed.AddField(fieldName, fieldValue, false)
			fieldCounter++
		}

		for _, r := range results {
			if fieldCounter >= maxFields {
				more = true
				break
			}
			title := r.Name
			if title == "" {
				title = fmt.Sprintf("ID: %d", r.Id)
			}
			if r.IsCategory {
				title += " (Category)"
			} else {
				title += " (Channel)"
			}
			fieldName := fmt.Sprintf("**%s**\n%s", title, separator)
			fieldValue := ""
			for _, missing := range r.Missing {
				fieldValue += fmt.Sprintf("**%s**\n%s\n", missing.Permission.String(), ctx.GetMessage(missing.Description))
			}
			sectionCount++
			if sectionCount < totalSections {
				fieldValue += blank
			}
			embed.AddField(fieldName, fieldValue, false)
			fieldCounter++
		}

		if more {
			embed.AddField("More...", "There are more missing permissions not shown due to Discord's 25 field limit.", false)
		}

		ctx.ReplyWithEmbed(embed)
	}
}
