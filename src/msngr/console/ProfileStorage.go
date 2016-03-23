package console

import (
	"database/sql"
	_ "github.com/lib/pq"
	"fmt"
	"strings"
	"log"
	"reflect"
)

type ProfileGroup struct {
	Name        string `json:"name"`
	Id          int64 `json:"id"`
	Description string `json:"description"`
}

type ProfileContact struct {
	ContactId   int64 `json:"id"`
	Address     string `json:"address"`
	Description string `json:"description"`
	Geo         Coordinates `json:"geo"`
	Links       []ProfileContactLink `json:"links"`
	OrderNumber int        `json:"order_number"`
}

func (pc ProfileContact) String() string {
	return fmt.Sprintf("\n\tContact [%v] position: %v\n\taddress: %v\n\tdescription: %v\n\tgeo: %+v\n\tlinks:%+v\n",
		pc.ContactId, pc.OrderNumber, pc.Address, pc.Description, pc.Geo, pc.Links,
	)
}

type ProfileContactLink struct {
	LinkId      int64  `json:"id"`
	Type        string `json:"type"`
	Value       string `json:"value"`
	Description string `json:"description"`
	OrderNumber int    `json:"order_number"`
}

func (pcl ProfileContactLink) String() string {
	return fmt.Sprintf("\n\t\tLink [%v] position: %v type: %v\n\t\tvalue: %v\n\t\tdescription: %v\n",
		pcl.LinkId, pcl.OrderNumber, pcl.Type, pcl.Value, pcl.Description,
	)
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

func (p *Profile) Equal(p1 *Profile) bool {
	return reflect.DeepEqual(p, p1)
}
func (p Profile) String() string {
	return fmt.Sprintf("\nPROFILE------------------\n: %v [%v] enable: %v, public: %v \nimg: %v\ndescriptions: %v %v \ncontacts: %+v \ngroups: %v \n----------------------\n",
		p.Name, p.UserName, p.Enable, p.Public, p.ImageURL, p.ShortDescription, p.TextDescription, p.Contacts, p.Groups,
	)
}
func NewProfileFromRow(row *sql.Rows) Profile {
	var id, short_text, long_text, image, name string
	var enable, public int
	err := row.Scan(&id, &short_text, &long_text, &image, &name, &enable, &public)
	if err != nil {
		log.Printf("P Error at scan profile data %v", err)
	}
	profile := Profile{UserName:id, ShortDescription:short_text, TextDescription:long_text, ImageURL:image, Name:name}
	if enable != 0 {
		profile.Enable = true
	}
	if public != 0 {
		profile.Public = true
	}
	return profile
}

type ProfileDbHandler struct {
	db *sql.DB
}

func (ph *ProfileDbHandler)FillProfileGroupsAndContact(profile *Profile) error {
	contacts, err := ph.GetProfileContacts(profile.UserName)
	if err != nil {
		log.Printf("P ERROR profile %v error load contacts", profile.UserName)
	}
	profile.Contacts = contacts

	groups, err := ph.GetProfileGroups(profile.UserName)
	if err != nil {
		log.Printf("P ERROR profile %v error load groups", profile.UserName)
	}
	profile.Groups = groups

	return nil
}

func NewProfileDbHandler(connectionString string) (*ProfileDbHandler, error) {
	pg, err := sql.Open("postgres", connectionString)
	if err != nil {
		log.Printf("CS Error at connect to db [%v]: %v", connectionString, err)
		return nil, err
	}
	ph := &ProfileDbHandler{db:pg}
	return ph, nil
}

func (ph *ProfileDbHandler) GetContactLinkTypes() []string {
	return []string{
		"phone", "WWW", "site",
	}
}
func (ph *ProfileDbHandler) GetProfileContacts(userName string) ([]ProfileContact, error) {
	contacts := []ProfileContact{}
	contactRows, err := ph.db.Query("SELECT pc.id, pc.address, pc.lat, pc.lon, pc.descr, pc.ord FROM profile_contacts pc WHERE pc.username = $1 ORDER BY pc.ord ASC", userName)
	if err != nil {
		log.Printf("P ERROR at query profile [%v] contacts %v", userName, err)
		return contacts, err
	}
	for contactRows.Next() {
		var cId int64
		var cOrd int
		var address string
		var descr sql.NullString
		var lat, lon float64
		err = contactRows.Scan(&cId, &address, &lat, &lon, &descr, &cOrd)
		if err != nil {
			log.Printf("P ERROR at scan profile [%v] contacts %v", userName, err)
			continue
		}
		var description string
		if descr.Valid {
			description = descr.String
		}
		contact := ProfileContact{ContactId:cId, Address:address, Geo:Coordinates{Lat:lat, Lon:lon}, OrderNumber:cOrd, Description:description}
		linkRows, err := ph.db.Query("SELECT l.id, l.ctype, l.cvalue, l.descr, l.ord FROM contact_links l WHERE l.contact_id = $1 ORDER BY l.ord ASC", cId)
		if err != nil {
			log.Printf("P ERROR at query to contact links [%+v] err: %v", contact, err)
			continue
		}
		for linkRows.Next() {
			var lType, lValue string
			var lId int64
			var lOrd int
			var lDescr sql.NullString
			err = linkRows.Scan(&lId, &lType, &lValue, &lDescr, &lOrd)
			if err != nil {
				log.Printf("P ERROR at scan contact link for contact_id = %v, %v", cId, err)
				continue
			}
			var lDescription string
			if lDescr.Valid{
				lDescription = lDescr.String
			}
			contactLink := ProfileContactLink{LinkId:lId, Type:lType, Description:lDescription, OrderNumber:lOrd, Value:lValue}
			contact.Links = append(contact.Links, contactLink)
		}

		contacts = append(contacts, contact)
	}
	return contacts, nil
}

func (ph *ProfileDbHandler) GetAllProfiles() ([]Profile, error) {
	profiles := []Profile{}
	profileRows, err := ph.db.Query("SELECT p.username, p.short_text, p.long_text, i.path, vs.fn, p.enable, p.public FROM profile p INNER JOIN profile_icons i ON p.username = i.username INNER JOIN vcard_search vs ON vs.username = p.username")
	if err != nil {
		log.Printf("P ERROR at query profiles: %v", err)
		return profiles, err
	}
	for profileRows.Next() {
		profile := NewProfileFromRow(profileRows)
		ph.FillProfileGroupsAndContact(&profile)
		profiles = append(profiles, profile)
	}
	return profiles, nil
}

func (ph *ProfileDbHandler) GetProfile(username string) (*Profile, error) {
	profileRow, err := ph.db.Query("SELECT p.username, p.short_text, p.long_text, i.path, vs.fn, p.enable, p.public FROM profile p INNER JOIN profile_icons i ON p.username = i.username INNER JOIN vcard_search vs ON vs.username = p.username WHERE p.username = $1", username)
	if err != nil {
		log.Printf("P ERROR at query profiles: %v", err)
		return nil, err
	}
	if profileRow.Next() {
		profile := NewProfileFromRow(profileRow)
		ph.FillProfileGroupsAndContact(&profile)
		return &profile, nil
	}
	return nil, nil
}

func (ph *ProfileDbHandler) InsertNewProfile(p *Profile) (*Profile, error) {
	err := ph.db.Ping()
	ph.db.QueryRow(fmt.Sprintf("INSERT INTO vcard (username, vcard) VALUES ('%v', '<vCard xmlns=''vcard-temp''><FN>%v</FN></vCard>');", p.UserName, p.Name))
	ph.db.QueryRow(fmt.Sprintf("INSERT INTO vcard_search(username, lusername, fn, lfn, family, lfamily, given, lgiven, middle, lmiddle, nickname, lnickname, bday, lbday, ctry, lctry, locality, llocality, email, lemail, orgname, lorgname, orgunit, lorgunit)  values ('%v', '%v', '%v', '%v', '', '', '', '', '', '', '', '', '', '', '', '', '', '', '', '', '', '', '', '');",
		p.UserName, strings.ToLower(p.UserName), p.Name, strings.ToLower(p.Name)))

	enable := 0
	if p.Enable {
		enable = 1
	}
	public := 0
	if p.Public {
		public = 1
	}
	ph.db.QueryRow(fmt.Sprintf("INSERT INTO profile (username, phonenumber, short_text, long_text, enable, public) VALUES ('%v', NULL, '%v', '%v', '%v', '%v');",
		p.UserName, p.ShortDescription, p.TextDescription, enable, public))
	ph.db.QueryRow(fmt.Sprintf("INSERT INTO profile_icons(username, path, itype) values('%v', '%v', 'profile');", p.UserName, p.ImageURL))

	for cInd, contact := range p.Contacts {
		log.Printf("P insert new profile [%v] add contact %+v", p.UserName, contact)
		if updContact, _ := ph.AddContactToProfile(p.UserName, &contact); updContact != nil {
			p.Contacts[cInd] = *updContact
		}
	}
	for gInd, group := range p.Groups {
		if updGroup, _ := ph.AddGroupToProfile(p.UserName, &group); updGroup != nil {
			p.Groups[gInd] = *updGroup
		}
	}
	return p, err
}
func (ph *ProfileDbHandler) BindGroupToProfile(userName string, group *ProfileGroup) error {
	_, err := ph.db.Exec("INSERT INTO profile_groups (username, group_id) VALUES ($1, $2)", userName, group.Id)
	if err != nil {
		//log.Printf("P ERROR at binding profile %v and group %+v: %v", userName, group, err)
		return err
	}
	return nil
}

func (ph *ProfileDbHandler) UnbindGroupsFromProfile(userName string) error {
	stmt, err := ph.db.Prepare("DELETE FROM profile_groups WHERE username=$1")
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(userName)
	if err != nil {
		return err
	}
	return nil
}

func (ph *ProfileDbHandler) InsertGroup(group *ProfileGroup) (*ProfileGroup, error) {
	var groupId int64
	err := ph.db.QueryRow("INSERT INTO groups (name, descr) VALUES ($1, $2) RETURNING id;", group.Name, group.Description).Scan(&groupId)
	if err != nil {
		log.Printf("P ERROR at inserting group %+v: %v", group, err)
		return nil, err
	}
	group.Id = groupId
	return group, nil
}
func (ph *ProfileDbHandler) AddGroupToProfile(userName string, group *ProfileGroup) (*ProfileGroup, error) {
	row, err := ph.db.Query("SELECT id FROM groups WHERE name=$1", group.Name)
	if err != nil {
		log.Printf("P ERROR add group to profile %v", err)
		return nil, err
	}
	if row.Next() {
		var gId int64
		row.Scan(&gId)
		group.Id = gId
	} else {
		group, err = ph.InsertGroup(group)
		if err != nil {
			return nil, err
		}
	}
	err = ph.BindGroupToProfile(userName, group)
	if err != nil {
		return nil, err
	}
	return group, nil
}

func (ph *ProfileDbHandler) GetProfileGroups(userName string) ([]ProfileGroup, error) {
	result := []ProfileGroup{}
	row, err := ph.db.Query("select g.id, g.name, g.descr from groups g inner join profile_groups pg on pg.group_id = g.id where pg.username=$1", userName)
	if err != nil {
		log.Printf("P ERROR at get profiles group for %v: %v", userName, err)
		return result, err
	}
	for row.Next() {
		var gId int64
		var name, descr string
		err = row.Scan(&gId, &name, &descr)
		if err != nil {
			log.Printf("P ERROR at get profiles group in scan: %v", err)
			continue
		}
		result = append(result, ProfileGroup{Id:gId, Name:name, Description:descr})
	}
	return result, nil
}

func (ph *ProfileDbHandler) InsertContact(userName string, contact *ProfileContact) (*ProfileContact, error) {
	var contactId int64
	err := ph.db.QueryRow("INSERT INTO profile_contacts (username, address, lat, lon, descr, ord) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id",
		userName, contact.Address, contact.Geo.Lat, contact.Geo.Lon, contact.Description, contact.OrderNumber).Scan(&contactId)
	if err != nil {
		log.Printf("P ERROR at add contact %+v to profile %v", contact, err)
		return nil, err
	}
	contact.ContactId = contactId
	return contact, nil
}

func (ph *ProfileDbHandler) AddContactToProfile(userName string, contact *ProfileContact) (*ProfileContact, error) {
	result, err := ph.InsertContact(userName, contact)
	if err != nil {
		return nil, err
	}
	for lInd, link := range result.Links {
		if updLink, err := ph.InsertContactLink(&link, contact.ContactId); updLink != nil {
			contact.Links[lInd] = *updLink
		} else {
			log.Printf("P ERROR at insert contact link %+v %v", link, err)
		}
	}
	return result, nil
}

func (ph *ProfileDbHandler) UpsertContact(userName string, newContact *ProfileContact) error {
	stmt, err := ph.db.Prepare("UPDATE profile_contacts SET address=$1, lat=$2, lon=$3, descr=$4, ord=$5 WHERE id=$6")
	defer stmt.Close()
	if err != nil {
		log.Printf("P ERROR at prepare update for change profile contact %v", err)
		return err
	}
	upd_res, err := stmt.Exec(newContact.Address, newContact.Geo.Lat, newContact.Geo.Lon, newContact.Description, newContact.OrderNumber, newContact.ContactId)
	if err != nil {
		log.Printf("P ERROR at execute update for change profile contact %v", err)
		return err
	}
	cRows, err := upd_res.RowsAffected()
	if err != nil {
		log.Printf("P ERROR at upsert contact in get rows update %v", err)
		return err
	}
	if cRows == 0 {
		log.Printf("P update contact of profile %v; add contact: %+v", userName, newContact)
		updatedContact, err := ph.InsertContact(userName, newContact)
		if err != nil {
			log.Printf("P ERROR at upsert contact in add contact to profile %v", err)
			return err
		}
		newContact.ContactId = updatedContact.ContactId
	}

	for _, link := range newContact.Links {
		if c, _ := ph.UpdateContactLink(link); c == 0 {
			ph.InsertContactLink(&link, newContact.ContactId)
		}
	}
	return nil
}

func (ph *ProfileDbHandler)DeleteContact(contactId int64) {
	ph.DeleteContactLinks(contactId)
	ph.deleteFromTable("profile_contacts", "id", contactId)
}

func (ph *ProfileDbHandler)InsertContactLink(link *ProfileContactLink, contactId int64) (*ProfileContactLink, error) {
	var lId int64
	err := ph.db.QueryRow("INSERT INTO contact_links (contact_id, ctype, cvalue, descr, ord) VALUES ($1, $2, $3, $4, $5) RETURNING id",
		contactId, link.Type, link.Value, link.Description, link.OrderNumber).Scan(&lId)
	if err != nil {
		log.Printf("P ERROR at insert contact link %v", err)
		return nil, err
	}
	log.Printf("P link id: %v", lId)
	link.LinkId = lId
	return link, nil
}

func (ph *ProfileDbHandler)UpdateContactLink(newLink ProfileContactLink) (int64, error) {
	stmt, err := ph.db.Prepare("UPDATE contact_links SET ctype=$1, cvalue=$2, descr=$3, ord=$4 WHERE id=$5")
	if err != nil {
		log.Printf("P ERROR at prepare update for change profile contact link %v", err)
		return -1, err
	}
	defer stmt.Close()
	upd_res, err := stmt.Exec(newLink.Type, newLink.Value, newLink.Description, newLink.OrderNumber, newLink.LinkId)
	if err != nil {
		log.Printf("P ERROR at execute update for change profile contact %v", err)
		return -1, err
	}
	countRows, err := upd_res.RowsAffected()
	if err != nil {
		return -1, err
	}
	return countRows, nil
}

func (ph *ProfileDbHandler)DeleteContactLinks(contactId int64) error {
	err := ph.deleteFromTable("contact_links", "contact_id", contactId)
	return err
}

func (ph *ProfileDbHandler)updateProfileField(tableName, fieldName, userName string, newValue interface{}) {
	stmt, err := ph.db.Prepare(fmt.Sprintf("UPDATE %v SET %v=$1 WHERE username=$2", tableName, fieldName))
	if err != nil {
		log.Printf("Error at prepare update for change profile [%v] %v %v", userName, fieldName, err)
	}
	defer stmt.Close()
	_, err = stmt.Exec(newValue, userName)
	if err != nil {
		log.Printf("Error at execute update for change profile [%v] %v %v", userName, fieldName, err)
	}
}

func (ph *ProfileDbHandler)deleteFromTable(tableName, nameId string, deleteId interface{}) error {
	stmt, err := ph.db.Prepare(fmt.Sprintf("DELETE FROM %v WHERE username=$1", tableName))
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(deleteId)
	if err != nil {
		return err
	}
	return nil
}

func (ph *ProfileDbHandler)DeleteProfile(userName string) error {
	//name
	ph.deleteFromTable("vcard", "username", userName)
	ph.deleteFromTable("vcard_search", "username", userName)
	//contacts
	contacts, _ := ph.GetProfileContacts(userName)
	for _, contact := range contacts {
		ph.DeleteContact(contact.ContactId)
	}
	ph.deleteFromTable("profile_contacts", "username", userName)
	//groups
	ph.UnbindGroupsFromProfile(userName)
	//data
	ph.deleteFromTable("profile", "username", userName)
	return nil
}

func (ph *ProfileDbHandler)UpdateProfile(newProfile *Profile) error {
	savedProfile, err := ph.GetProfile(newProfile.UserName)
	if err != nil {
		return err
	}
	if savedProfile == nil {
		ph.InsertNewProfile(newProfile)
		return nil
	}

	if savedProfile.Enable != newProfile.Enable {
		enable := 0
		if newProfile.Enable {
			enable = 1
		}
		ph.updateProfileField("profile", "enable", newProfile.UserName, enable)
	}
	if savedProfile.Public != newProfile.Public {
		public := 0
		if newProfile.Public {
			public = 1
		}
		ph.updateProfileField("profile", "public", newProfile.UserName, public)
	}
	if savedProfile.ImageURL != newProfile.ImageURL {
		ph.updateProfileField("profile_icons", "path", newProfile.UserName, newProfile.ImageURL)
	}
	if savedProfile.Name != newProfile.Name {
		stmt, err := ph.db.Prepare(fmt.Sprintf("UPDATE vcard SET vcard='<vCard xmlns=''vcard-temp''><FN>%v</FN></vCard>' WHERE username=$1", newProfile.Name))
		defer stmt.Close()
		if err != nil {
			log.Printf("Error at prepare update for change profile [%v] public %v", newProfile.UserName, err)
		}
		_, err = stmt.Exec(newProfile.UserName)
		if err != nil {
			log.Printf("Error at execute update for change profile [%v] public %v", newProfile.UserName, err)
		}
		stmt_s, err := ph.db.Prepare("UPDATE vcard_search SET fn=$1, lfn=$2 WHERE username=$3")
		if err != nil {
			log.Printf("Error at prepare update for change profile [%v] public %v", newProfile.UserName, err)
		}
		defer stmt_s.Close()
		_, err = stmt_s.Exec(newProfile.Name, strings.ToLower(newProfile.Name), newProfile.UserName)
		if err != nil {
			log.Printf("Error at execute update for change profile [%v] public %v", newProfile.UserName, err)
		}
	}

	if savedProfile.ShortDescription != newProfile.ShortDescription {
		ph.updateProfileField("profile", "short_text", newProfile.UserName, newProfile.ShortDescription)
	}

	if savedProfile.TextDescription != newProfile.TextDescription {
		ph.updateProfileField("profile", "long_text", newProfile.UserName, newProfile.TextDescription)
	}

	if !reflect.DeepEqual(savedProfile.Contacts, newProfile.Contacts) {
		log.Printf("Difference in contacts")
		for _, contact := range newProfile.Contacts {
			ph.UpsertContact(newProfile.UserName, &contact)
		}
	}
	if !reflect.DeepEqual(savedProfile.Groups, newProfile.Groups) {
		log.Printf("Difference in groups")
		ph.UnbindGroupsFromProfile(newProfile.UserName)
		for _, group := range newProfile.Groups {
			ph.AddGroupToProfile(newProfile.UserName, &group)
		}
	}
	return nil
}

