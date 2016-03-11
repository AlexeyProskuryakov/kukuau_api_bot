package console

import (
	"database/sql"
	"fmt"
	"strings"
	"log"
	"reflect"
)

type ProfileGroup struct {
	GroupName string `json:"group_name"`
	GroupId   int64 `json:"group_id"`
}

type ProfileContact struct {
	ContactId   string `json:"contact_id"`
	Address     string `json:"address"`
	Description string `json:"description"`
	Geo         Coordinates `json:"place"`
	Links       []ProfileContactLink `json:"links"`
	OrderNumber int        `json:"order_number"`
}

type ProfileContactLink struct {
	LinkId      int64  `json:"link_id"`
	Type        string `json:"type"`
	Value       string `json:"value"`
	Description string `json:"description"`
	OrderNumber int    `json:"order_number"`
}

type Coordinates struct {
	Lat float64 `json:"lat"`
	Lon float64 `json:"lon"`
}

type Profile struct {
	UserName         string `json:"id"`
	ImageURL         string `json:"image_url"`
	Name             string `json:"name"`
	ShortDescription string `json:"short_description"`
	TextDescription  string `json:"text_description"`
	Contacts         []ProfileContact `json:"contacts"`
	Groups           []ProfileGroup `json:"groups"`
	Enable           bool `json:"enable"`
	Public           bool `json:"public"`
}

func GetProfileContacts(db *sql.DB, userName string) ([]ProfileContact, error) {
	contacts := []ProfileContact{}
	contactRows, err := db.Query("SELECT pc.id, pc.address, pc.lat, pc.lon, pc.ord FROM profile_contacts pc WHERE pc.username = $1", userName)
	if err != nil {
		log.Printf("CS Error at query profile [%v] contacts %v", userName, err)
		return contacts, err
	}
	for contactRows.Next() {
		var cId int64
		var cOrd int
		var address string
		var lat, lon float64
		err = contactRows.Scan(&cId, &address, &lat, &lon, &cOrd)
		if err != nil {
			log.Printf("ERROR at scan profile [%v] contacts %v", userName, err)
			continue
		}
		contact := ProfileContact{ContactId:cId, Address:address, Geo:Coordinates{Lat:lat, Lon:lon}, OrderNumber:cOrd}
		linkRows, err := db.Query("SELECT l.id, l.ctype, l.cvalue, l.descr, l.ord FROM contact_links l WHERE l.contact_id = $1", cId)
		if err != nil {
			log.Printf("ERROR at query to contact links [%+v]", contact)
			continue
		}
		for linkRows.Next() {
			var lType, lValue, lDescr string
			var lId int64
			var lOrd int
			err = linkRows.Scan(&lId, &lType, &lValue, &lDescr, &lOrd)
			if err != nil {
				log.Printf("ERROR at scan contact link")
				continue
			}
			contactLink := ProfileContactLink{LinkId:lId, Type:lType, Description:lDescr, OrderNumber:lOrd}
			contact.Links = append(contact.Links, contactLink)
		}

		contacts = append(contacts, contact)
		log.Printf("CS Add to profile [%v] contacts %+v", userName, contacts)
	}
	return contacts, nil
}
func GetAllProfiles(db *sql.DB) ([]Profile, error) {
	profiles := []Profile{}
	profileRows, err := db.Query("SELECT p.username, p.short_text, p.long_text, i.path, vs.fn, p.enable, p.public FROM profile p INNER JOIN profile_icons i ON p.username = i.username INNER JOIN vcard_search vs ON vs.username = p.username")
	if err != nil {
		log.Printf("CS Error at query profiles: %v", err)
		return profiles, err
	}
	for profileRows.Next() {
		var id, short_text, long_text, image, name string
		var enable, public int
		err = profileRows.Scan(&id, &short_text, &long_text, &image, &name, &enable, &public)
		if err != nil {
			log.Printf("CS Error at scan profile data %v", err)
		}
		profile := Profile{UserName:id, ShortDescription:short_text, TextDescription:long_text, ImageURL:image, Name:name}
		if enable != 0 {
			profile.Enable = true
		}
		if public != 0 {
			profile.Public = true
		}
		contacts, err := GetProfileContacts(db, profile.UserName)
		if err != nil {
			log.Printf("ERROR profile %v error load contacts", profile.UserName)
			continue
		}
		profile.Contacts = contacts
		profiles = append(profiles, profile)
	}
	return profiles, nil
}

func GetProfile(db *sql.DB, username string) (*Profile, error) {
	profileRow := db.QueryRow("SELECT p.username, p.short_text, p.long_text, i.path, vs.fn, p.enable, p.public FROM profile p INNER JOIN profile_icons i ON p.username = i.username INNER JOIN vcard_search vs ON vs.username = p.username WHERE p.username = $1", username)
	var id, short_text, long_text, image, name string
	var enable, public int
	err := profileRow.Scan(&id, &short_text, &long_text, &image, &name, &enable, &public)
	if err != nil {
		log.Printf("CS Error at scan profile data %v", err)
	}
	profile := Profile{UserName:id, ShortDescription:short_text, TextDescription:long_text, ImageURL:image, Name:name}
	if enable != 0 {
		profile.Enable = true
	}
	if public != 0 {
		profile.Public = true
	}
	contacts, err := GetProfileContacts(db, profile.UserName)
	if err != nil {
		log.Printf("ERROR profile %v error load contacts", profile.UserName)
	}
	profile.Contacts = contacts
	return &profile, nil
}

func InsertNewProfile(db *sql.DB, p Profile) {
	db.QueryRow(fmt.Sprintf("INSERT INTO vcard (username, vcard) VALUES ('%v', '<vCard xmlns=''vcard-temp''><FN>%v</FN></vCard>');", p.UserName, p.Name))
	db.QueryRow(fmt.Sprintf("INSERT INTO vcard_search(username, lusername, fn, lfn, family, lfamily, given, lgiven, middle, lmiddle, nickname, lnickname, bday, lbday, ctry, lctry, locality, llocality, email, lemail, orgname, lorgname, orgunit, lorgunit)  values ('%v', '', '%v', '%v', '', '', '', '', '', '', '', '', '', '', '', '', '', '', '', '', '', '', '', '');",
		p.UserName, p.Name, strings.ToLower(p.Name)))

	enable := 0
	if p.Enable {
		enable = 1
	}
	public := 0
	if p.Public {
		public = 1
	}
	db.QueryRow(fmt.Sprintf("INSERT INTO profile (username, phonenumber, short_text, long_text, enable, public) VALUES ('%v', NULL, '%v', '%v', '%v', '%v');",
		p.UserName, p.ShortDescription, p.TextDescription, enable, public))
	db.QueryRow(fmt.Sprintf("INSERT INTO profile_icons(username, path, itype) values('%v', '%v', 'profile');", p.UserName, p.ImageURL))

	for _, contact := range p.Contacts {
		err := InsertContact(db, contact, p.UserName)
		if err != nil {
			log.Printf("P ERROR at insert contact %+v", contact)
			continue
		}

	}
}

func InsertContact(db *sql.DB, contact ProfileContact, userName string) error {
	res, err := db.Exec("INSERT INTO profile_contacts (username, address, lat, lon, descr, ord) VALUES ($1, $2, $3, $4, $5, $6);",
		userName, contact.Address, contact.Geo.Lat, contact.Geo.Lon, contact.OrderNumber)
	log.Printf("P INSERT contact %v", res.RowsAffected())
	for _, link := range contact.Links {
		err = InsertContactLink(db, link, contact.ContactId)
		if err != nil {
			log.Printf("P ERROR at insert contact link %+v", link)
		}

	}
	return err
}

func UpdateContact(db *sql.DB, newContact ProfileContact, userName string) (int, error) {
	stmt, err := db.Prepare("UPDATE profile_contacts SET address=$1 lat=$2 lon=$3 descr=$4 ord=$5 WHERE id=$6")
	defer stmt.Close()
	if err != nil {
		log.Printf("P ERROR at prepare update for change profile contact %v", err)
		return -1, err
	}
	upd_res, err := stmt.Exec(newContact.Address, newContact.Geo.Lat, newContact.Geo.Lon, newContact.Description, newContact.OrderNumber, userName)
	if err != nil {
		log.Printf("P ERROR at execute update for change profile contact %v", err)
		return -1, err
	}
	for _, link := range newContact.Links {
		c, _ := UpdateContactLink(db, link)
		if c == 0 {
			InsertContactLink(db, link, newContact.ContactId)
		}
	}
	return upd_res.RowsAffected()
}

func DeleteContact(db *sql.DB, contactId int64) {
	DeleteContactLinks(db, contactId)
	deleteFromTable(db, "profile_contacts", "id", contactId)
}

func InsertContactLink(db *sql.DB, link ProfileContactLink, contactId int64) error {
	res, err := db.Exec("INSERT INTO contact_links (contact_id, ctype, cvalue, descr, ord) VALUES ($1, $2, $3, $4, $5);",
		contactId, link.Type, link.Value, link.Description, link.OrderNumber)
	log.Printf("P INSERT contact link %v", res.RowsAffected())
	return err
}

func UpdateContactLink(db *sql.DB, newLink ProfileContactLink) (int, error) {
	stmt, err := db.Prepare("UPDATE contact_links SET ctype=$1 cvalue=$2 descr=$3 ord=$4 WHERE id=$5")
	defer stmt.Close()
	if err != nil {
		log.Printf("P ERROR at prepare update for change profile contact link %v", err)
		return -1, err
	}
	upd_res, err := stmt.Exec(newLink.Type, newLink.Value, newLink.Description, newLink.OrderNumber, newLink.LinkId)
	if err != nil {
		log.Printf("P ERROR at execute update for change profile contact %v", err)
		return -1, err
	}
	return upd_res.RowsAffected()
}

func DeleteContactLinks(db *sql.DB, contactId int64) error {
	err := deleteFromTable(db, "contact_links", "contact_id", contactId)
	return err
}

func updateProfileField(db *sql.DB, tableName, fieldName, userName string, newValue interface{}) {
	stmt, err := db.Prepare(fmt.Sprintf("UPDATE %v SET %v=$1 WHERE username=$2", tableName, fieldName))
	defer stmt.Close()
	if err != nil {
		log.Printf("Error at prepare update for change profile [%v] %v %v", userName, fieldName, err)
	}
	_, err = stmt.Exec(newValue, userName)
	if err != nil {
		log.Printf("Error at execute update for change profile [%v] %v %v", userName, fieldName, err)
	}
}

func deleteFromTable(db *sql.DB, tableName, nameId string, deleteId interface{}) error {
	stmt, err := db.Prepare(fmt.Sprintf("DELETE FROM %v WHERE username=$1", tableName))
	defer stmt.Close()
	if err != nil {
		return err
	}
	_, err = stmt.Exec(deleteId)
	if err != nil {
		return err
	}
	return nil
}

func DeleteProfile(db *sql.DB, userName string) error {
	deleteFromTable(db, "profile", "username", userName)
	deleteFromTable(db, "vcard", "username", userName)
	deleteFromTable(db, "vcard_search", "username", userName)
	contacts, _ := GetProfileContacts(db, userName)
	for _, contact := range contacts{
		DeleteContact(db, contact.ContactId)
	}
	deleteFromTable(db, "profile_contacts", "username", userName)
	return nil
}

func UpdateProfile(db *sql.DB, newProfile Profile) error {
	savedProfile, err := GetProfile(db, newProfile.UserName)
	if err != nil {
		return err
	}
	if savedProfile.Enable != newProfile.Enable {
		enable := 0
		if newProfile.Enable {
			enable = 1
		}
		updateProfileField(db, "profile", "enable", newProfile.UserName, enable)
	}
	if savedProfile.Public != newProfile.Public {
		public := 0
		if newProfile.Public {
			public = 1
		}
		updateProfileField(db, "profile", "public", newProfile.UserName, public)
	}
	if savedProfile.ImageURL != newProfile.ImageURL {
		updateProfileField(db, "profile_icons", "path", newProfile.UserName, newProfile.ImageURL)
	}
	if savedProfile.Name != newProfile.Name {
		stmt, err := db.Prepare(fmt.Sprintf("UPDATE vcard SET vcard='<vCard xmlns=''vcard-temp''><FN>%v</FN></vCard>' WHERE username=$1", newProfile.Name))
		defer stmt.Close()
		if err != nil {
			log.Printf("Error at prepare update for change profile [%v] public %v", newProfile.UserName, err)
		}
		_, err = stmt.Exec(newProfile.UserName)
		if err != nil {
			log.Printf("Error at execute update for change profile [%v] public %v", newProfile.UserName, err)
		}
		stmt_s, err := db.Prepare("UPDATE vcard_search SET fn=$1 lfn=$2 WHERE username=$3")
		defer stmt_s.Close()
		if err != nil {
			log.Printf("Error at prepare update for change profile [%v] public %v", newProfile.UserName, err)
		}
		_, err = stmt_s.Exec(newProfile.Name, strings.ToLower(newProfile.Name), newProfile.UserName)
		if err != nil {
			log.Printf("Error at execute update for change profile [%v] public %v", newProfile.UserName, err)
		}
	}

	if savedProfile.ShortDescription != newProfile.ShortDescription {
		updateProfileField(db, "profile", "short_text", newProfile.UserName, newProfile.ShortDescription)
	}

	if savedProfile.TextDescription != newProfile.TextDescription {
		updateProfileField(db, "profile", "long_text", newProfile.UserName, newProfile.TextDescription)
	}

	if !reflect.DeepEqual(savedProfile.Contacts, newProfile.Contacts) {
		for _, contact := range newProfile.Contacts {
			c, _ := UpdateContact(db, contact, newProfile.UserName)
			if c == 0 {
				InsertContact(db, contact, newProfile.UserName)
			}
		}
	}
	return nil
}

