package utils

import (
	"github.com/dorkitude/linctl/pkg/oauth"
)

// ActorParams represents actor attribution parameters
type ActorParams struct {
	Actor     string
	AvatarURL string
}

// ResolveActorParams resolves actor parameters using provided values and environment defaults
func ResolveActorParams(providedActor, providedAvatarURL string) *ActorParams {
	// Load actor configuration from environment
	actorConfig := oauth.LoadActorFromEnvironment()

	return &ActorParams{
		Actor:     actorConfig.GetActor(providedActor),
		AvatarURL: actorConfig.GetAvatarURL(providedAvatarURL),
	}
}

// HasActorInfo returns true if any actor information is available
func (ap *ActorParams) HasActorInfo() bool {
	return ap != nil && (ap.Actor != "" || ap.AvatarURL != "")
}

// ToCreateAsUser returns the actor name for createAsUser field, or nil if empty
func (ap *ActorParams) ToCreateAsUser() *string {
	if ap == nil || ap.Actor == "" {
		return nil
	}
	return &ap.Actor
}

// ToDisplayIconURL returns the avatar URL for displayIconUrl field, or nil if empty
func (ap *ActorParams) ToDisplayIconURL() *string {
	if ap == nil || ap.AvatarURL == "" {
		return nil
	}
	return &ap.AvatarURL
}
