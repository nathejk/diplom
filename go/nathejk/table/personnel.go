package table

import (
	"fmt"
	"log"
	"time"

	"github.com/nathejk/shared-go/messages"
	"github.com/nathejk/shared-go/types"
	"nathejk.dk/pkg/tablerow"
	"nathejk.dk/superfluids/entity"
	"nathejk.dk/superfluids/streaminterface"

	_ "embed"
)

type Personnel struct {
	UserID     types.UserID
	Name       string
	Email      types.EmailAddress
	Phone      types.PhoneNumber
	MedlemNr   string
	Corps      types.CorpsSlug
	Department string
	HqAccess   bool
	Created    time.Time
	Updated    time.Time
}
type PersonnelTableEvent struct {
	UserID types.UserID
}

type personnel struct {
	w tablerow.Consumer
	p entity.EntityChangedPublisher
}

func NewPersonnel(w tablerow.Consumer, p streaminterface.Publisher) *personnel {
	table := &personnel{w: w, p: entity.NewEntityChangedPublisher(p, "personnel")}
	if err := w.Consume(table.CreateTableSql()); err != nil {
		log.Fatalf("Error creating table %q", err)
	}
	return table
}

//go:embed personnel.sql
var personnelSchema string

func (t *personnel) CreateTableSql() string {
	return personnelSchema
}

func (c *personnel) Consumes() (subjs []streaminterface.Subject) {
	return []streaminterface.Subject{
		streaminterface.SubjectFromStr("nathejk"),
	}
}

func (c *personnel) HandleMessage(msg streaminterface.Message) error {
	switch msg.Subject().Subject() {
	case "nathejk:personnel.updated":
		var body messages.NathejkPersonnelUpdated
		if err := msg.Body(&body); err != nil {
			return err
		}
		hqAccess := "0"
		if body.HqAccess {
			hqAccess = "1"
		}
		if body.Name == "" {
			err := c.w.Consume(fmt.Sprintf("INSERT INTO personnel (userId, phone, pincode, createdAt, updatedAt) VALUES (%q,%q,%q,%q,%q) ON DUPLICATE KEY UPDATE  phone=VALUES(phone), pincode=VALUES(pincode), updatedAt=VALUES(updatedAt)", body.UserID, body.Phone.Normalize(), body.Pincode, msg.Time(), msg.Time()))
			if err != nil {
				log.Fatalf("Error consuming sql %q", err)
			}
		} else {
			err := c.w.Consume(fmt.Sprintf("INSERT INTO personnel (userId, name,  email, phone, medlemNr, corps, department, hqAccess, diet, createdAt, updatedAt) VALUES (%q,%q,%q,%q,%q,%q,%q,%q,%q,%q,%q) ON DUPLICATE KEY UPDATE  name=VALUES(name), email=VALUES(email),phone=VALUES(phone), medlemNr=VALUES(medlemNr), corps=VALUES(corps), department=VALUES(department), hqAccess=VALUES(hqAccess), diet=VALUES(diet), updatedAt=VALUES(updatedAt)", body.UserID, body.Name, body.Email, body.Phone.Normalize(), body.MedlemNr, string(body.Corps), body.Department, hqAccess, body.Diet, msg.Time(), msg.Time()))
			if err != nil {
				log.Fatalf("Error consuming sql %q", err)
			}
		}
		c.p.Changed(&PersonnelTableEvent{UserID: body.UserID})

	case "nathejk:personnel.deleted":
		var body messages.NathejkPersonnelDeleted
		if err := msg.Body(&body); err != nil {
			return err
		}
		err := c.w.Consume(fmt.Sprintf("DELETE FROM personnel WHERE userId=%q", body.UserID))
		if err != nil {
			log.Fatalf("Error consuming sql %q", err)
		}
		c.p.Deleted(&PersonnelTableEvent{UserID: body.UserID})

	}
	return nil
}
