package services

import (
	"fmt"

	"uuid"

	"github.com/zaidmasri/business-planning-tool/internal/application/interfaces"
	"github.com/zaidmasri/business-planning-tool/internal/domain/ports"
	"github.com/zaidmasri/business-planning-tool/internal/domain/repositories"
)

type notificationService struct {
	mailer     ports.Mailer
	users      repositories.UserRepository
	plans      repositories.PlanRepository
	invites    repositories.InviteRepository
	appBaseURL string
}

func NewNotificationService(
	mailer ports.Mailer,
	users repositories.UserRepository,
	plans repositories.PlanRepository,
	invites repositories.InviteRepository,
	appBaseURL string,
) interfaces.NotificationService {
	return &notificationService{
		mailer:     mailer,
		users:      users,
		plans:      plans,
		invites:    invites,
		appBaseURL: appBaseURL,
	}
}

func (s *notificationService) SendWelcomeEmail(userID uuid.UUID) error {
	user, err := s.users.GetUser(userID)
	if err != nil {
		return fmt.Errorf("failed to load user for welcome email: %w", err)
	}

	subject := "Welcome to Northbasis"
	body := fmt.Sprintf(
		"Hi %s,\n\nThanks for signing up for Northbasis! You can start building your financial projections here:\n\n%s\n",
		greetingName(user.FirstName()), s.appBaseURL,
	)

	return s.mailer.Send(user.Email(), subject, body)
}

func (s *notificationService) SendInviteEmail(inviteID uuid.UUID) error {
	invite, err := s.invites.GetInvite(inviteID)
	if err != nil {
		return fmt.Errorf("failed to load invite for invite email: %w", err)
	}

	plan, err := s.plans.Get(invite.PlanID)
	if err != nil {
		return fmt.Errorf("failed to load plan for invite email: %w", err)
	}

	inviterName := "Someone"
	if inviter, err := s.users.GetUser(invite.InvitedBy); err == nil {
		inviterName = greetingName(inviter.FirstName())
	}

	subject := fmt.Sprintf("%s invited you to collaborate on %s", inviterName, plan.Name())
	body := fmt.Sprintf(
		"Hi,\n\n%s has invited you to collaborate on \"%s\" as %s %s.\n\nLog in or sign up at %s with this email address (%s) to see the invite.\n",
		inviterName, plan.Name(), article(string(invite.AccessLevel)), invite.AccessLevel, s.appBaseURL, invite.Email,
	)

	return s.mailer.Send(invite.Email, subject, body)
}

func greetingName(firstName string) string {
	if firstName == "" {
		return "there"
	}
	return firstName
}

// article returns "an" for a word starting with a vowel sound, "a" otherwise
// — good enough for the fixed set of access-level words this email uses
// (viewer, editor, owner).
func article(word string) string {
	if word == "" {
		return "a"
	}
	switch word[0] {
	case 'a', 'e', 'i', 'o', 'u', 'A', 'E', 'I', 'O', 'U':
		return "an"
	default:
		return "a"
	}
}
