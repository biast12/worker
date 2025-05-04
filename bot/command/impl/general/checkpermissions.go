package general

import (
	"fmt"
	"sort"
	"strings"
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
	Permission permissiongdl.Permission
	Description  i18n.MessageId
}

func (CheckPermissionsCommand) Execute(ctx registry.CommandContext) {
	worker := ctx.Worker()
	guildId := ctx.GuildId()
	botId := worker.BotId

	// Check for Administrator permission
	if permissionwrapper.HasPermissions(worker, guildId, botId, permissiongdl.Administrator) {
		ctx.ReplyRaw(customisation.Green, fmt.Sprintf("%s", i18n.PermissionsTitle), fmt.Sprintf("✅ <@%d> %s", botId, i18n.PermissionsHasAdministrator))
		return
	}

	permissionsToCheck := []PermissionCheck{
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

	// Check bot permissions
	missingBotPermissions := []string{}
	for _, check := range permissionsToCheck {
		if !permissionwrapper.HasPermissions(worker, guildId, botId, check.Permission) {
			missingBotPermissions = append(missingBotPermissions, string(check.Description))
		}
	}

	sort.Strings(missingBotPermissions)

	// Build response
	if len(missingBotPermissions) == 0 {
		ctx.ReplyRaw(customisation.Green, fmt.Sprintf("%s", i18n.PermissionsTitle), fmt.Sprintf("✅ <@%d> %s", botId, i18n.PermissionsHasAllPermissions))
	} else {
		embed := embed.NewEmbed().
			SetTitle(fmt.Sprintf("❌ %s", i18n.PermissionsMissing)).
			SetColor(int(customisation.Red)).
			SetDescription(strings.Join(missingBotPermissions, "\n"))
		ctx.ReplyWithEmbed(embed)
	}
}
