package node

import (
	"context"

	"github.com/pkg/errors"

	"github.com/berty/berty/core/api/p2p"
	"github.com/berty/berty/core/entity"
	"github.com/berty/berty/core/sql"
)

type EventHandler func(context.Context, *p2p.Event) error

func (n *Node) handleContactRequest(ctx context.Context, input *p2p.Event) error {
	attrs, err := input.GetContactRequestAttrs()
	if err != nil {
		return err
	}
	// FIXME: validate input

	// FIXME: check if contact is not already known

	// save requester in db
	requester := attrs.Me
	requester.Status = entity.Contact_RequestedMe
	requester.Devices = []*entity.Device{
		{
			ID: p2p.GetSender(ctx),
			//Key: crypto.NewPublicKey(p2p.GetPubkey(ctx)),
		},
	}
	if err := n.sql.Set("gorm:association_autoupdate", true).Save(requester).Error; err != nil {
		return err
	}

	// nothing more to do, now we wait for the UI to accept the request
	return nil
}

func (n *Node) handleContactRequestAccepted(ctx context.Context, input *p2p.Event) error {
	// fetching existing contact from db
	contact, err := sql.ContactByID(n.sql, p2p.GetSender(ctx))
	if err != nil {
		return errors.Wrap(err, "no such contact")
	}

	contact.Status = entity.Contact_IsFriend
	//contact.Devices[0].Key = crypto.NewPublicKey(p2p.GetPubkey(ctx))
	if err := n.sql.Set("gorm:association_autoupdate", true).Save(contact).Error; err != nil {
		return err
	}

	// send my contact
	if err := n.contactShareMe(contact); err != nil {
		return err
	}

	return nil
}

func (n *Node) handleContactShareMe(ctx context.Context, input *p2p.Event) error {
	attrs, err := input.GetContactShareMeAttrs()
	if err != nil {
		return err
	}

	// fetching existing contact from db
	contact, err := sql.ContactByID(n.sql, p2p.GetSender(ctx))
	if err != nil {
		return errors.Wrap(err, "no such contact")
	}

	// FIXME: UI: ask for confirmation before update
	contact.DisplayName = attrs.Me.DisplayName
	contact.DisplayStatus = attrs.Me.DisplayStatus
	// FIXME: save more attributes
	return n.sql.Save(contact).Error
}