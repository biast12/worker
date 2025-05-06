package general

import (
	"time"

	"github.com/TicketsBot-cloud/common/permission"
	"github.com/TicketsBot-cloud/worker/bot/command"
	"github.com/TicketsBot-cloud/worker/bot/command/registry"
	"github.com/TicketsBot-cloud/worker/bot/customisation"
	"github.com/TicketsBot-cloud/worker/bot/permissionwrapper"
	"github.com/TicketsBot-cloud/worker/i18n"
	"github.com/rxdn/gdl/objects/channel/embed"
	"github.com/rxdn/gdl/objects/interaction"
	permissiongdl "github.com/rxdn/gdl/permission"
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

func getPermissionsToCheck() []PermissionCheck {
	return []PermissionCheck{
		{permissiongdl.ViewChannel, i18n.PermissionsViewChannel},
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
}

func (CheckPermissionsCommand) Execute(ctx registry.CommandContext) {
	worker := ctx.Worker()
	guildId := ctx.GuildId()
	botId := worker.BotId

	// Check for Administrator permission
	if permissionwrapper.HasPermissions(worker, guildId, botId, permissiongdl.Administrator) {
		ctx.Reply(customisation.Green, i18n.PermissionsTitle, i18n.PermissionsHasAdministrator)
		return
	}

	permissionsToCheck := getPermissionsToCheck()

	// Check bot permissions
	missingBotPermissions := []PermissionCheck{}
	for _, check := range permissionsToCheck {
		if !permissionwrapper.HasPermissions(worker, guildId, botId, check.Permission) {
			missingBotPermissions = append(missingBotPermissions, check)
		}
	}

	// Build response
	if len(missingBotPermissions) == 0 {
		ctx.Reply(customisation.Green, i18n.PermissionsTitle, i18n.PermissionsHasAllPermissions, guildId)
	} else {
		embed := embed.NewEmbed().
			SetTitle(ctx.GetMessage(i18n.PermissionsMissing)).
			SetColor(ctx.GetColour(customisation.Red))

		for _, missing := range missingBotPermissions {
			embed.AddField(missing.Permission.String(), ctx.GetMessage(missing.Description), false)
		}

		ctx.ReplyWithEmbed(embed)
	}
}
