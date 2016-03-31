package console

import (
	"testing"
	"msngr/configuration"
)

var (
	config = configuration.ReadTestConfigInRecursive()

	profile1 = Profile{UserName:"test1", Name:"testProfile1", ImageURL:"http://www.foo.bar", ShortDescription:"test", TextDescription:"ffffffffffffffffffffffffffffff", }
	profile2 = Profile{UserName:"test2", Name:"testProfile2", ImageURL:"http://www.foo.bar", ShortDescription:"test", TextDescription:"ffffffffffffffffffffffffffffff", }
	profile3 = Profile{UserName:"test3", Name:"testProfile3", ImageURL:"http://www.foo.bar", ShortDescription:"test", TextDescription:"ffffffffffffffffffffffffffffff", }
	profileEnabled = Profile{UserName:"testEnabled", Name:"testEnabled", ImageURL:"http://www.foo.bar", ShortDescription:"test", TextDescription:"ffffffffffffffffffffffffffffff", Enable:true}
	profilePublic = Profile{UserName:"testPublic", Name:"testPublic", ImageURL:"http://www.foo.bar", ShortDescription:"test", TextDescription:"ffffffffffffffffffffffffffffff", Public:true}

	group = ProfileGroup{
		Name:"test_group",
		Description:"test_group_description",
	}
	contact = ProfileContact{
		Address:"test adress",
		Description:"test descr",
		OrderNumber:1,
		Lat:1,
		Lon:2,
		Links:[]ProfileContactLink{
			ProfileContactLink{Type:"www", Value:"http://", Description:"tututut", OrderNumber:1},
			ProfileContactLink{Type:"phone", Value:"+79811064022", Description:"tututut", OrderNumber:2},
		},
	}
	profileWithContacts = Profile{UserName:"test_w_c", Name:"with contact", Contacts:[]ProfileContact{contact}}
	profileWithGroup = Profile{UserName:"test_w_g", Name:"with group", Groups:[]ProfileGroup{group}}
)

func check_err(t *testing.T, err error, i string) {
	if err != nil {
		t.Errorf("ERROR: [%v]\n%v", i, err)
	}
}

func deleteAll(ph *ProfileDbHandler) {
	profiles, _ := ph.GetAllProfiles()
	for _, profile := range profiles {
		ph.DeleteProfile(profile.UserName)
	}
	stmt, _ := ph.db.Prepare("DELETE FROM groups")
	defer stmt.Close()
	stmt.Exec()

}

func testProfilesCount(t *testing.T, ph *ProfileDbHandler, count int) {
	profiles, err := ph.GetAllProfiles()
	if err != nil {
		t.Error(err)
	}
	if len(profiles) != count {
		t.Errorf("at db must be %v profiles but %v", count, len(profiles))
	}
}

func check_groups(t *testing.T, profile *Profile, count int) {
	if len(profile.Groups) != count {
		t.Errorf("Groups must be %v but %v", count, len(profile.Groups))
	}
	for _, g := range profile.Groups {
		if g.Id == 0 {
			t.Errorf("At profile: %+v\ngroup: %+v\nhave not id :(", profile, group)
		}
	}
}

func check_contacts(t *testing.T, profile *Profile) bool {
	for _, cnt := range profile.Contacts {
		if cnt.ContactId == 0 {
			t.Error("contact have not id", cnt)
			return false
		}
		for _, cntl := range cnt.Links {
			if cntl.LinkId == 0 {
				t.Error("contact link have not id", cntl)
				return false
			}
		}
	}
	return true
}

func TestProfilesCounts(t *testing.T) {
	t.SkipNow()
	ph, err := NewProfileDbHandler(config.Main.PGDatabase.ConnString)
	check_err(t, err, "init")
	deleteAll(ph)
	testProfilesCount(t, ph, 0)
	ph.InsertNewProfile(&profile1)
	testProfilesCount(t, ph, 1)

	ph.DeleteProfile(profile1.UserName)
	testProfilesCount(t, ph, 0)
	//
	ph.InsertNewProfile(&profile1)
	ph.InsertNewProfile(&profile2)
	ph.InsertNewProfile(&profile3)

	testProfilesCount(t, ph, 3)

	profile1 = Profile{UserName:"test1", Name:"testProfile1", ImageURL:"http://www.foo.bar", ShortDescription:"test", TextDescription:"ffffffffffffffffffffffffffffff", }
	profile1.Contacts = append(profile1.Contacts, ProfileContact{})
	profile1.Groups = append(profile1.Groups, ProfileGroup{Name:"fooo"})

	ph.UpdateProfile(&profile1)
	testProfilesCount(t, ph, 3)
}

func TestElements(t *testing.T) {
	t.SkipNow()
	ph, err := NewProfileDbHandler(config.Main.PGDatabase.ConnString)
	check_err(t, err, "init")
	deleteAll(ph)

	inserted_profile, _ := ph.InsertNewProfile(&profile1)
	if !inserted_profile.Equal(&profile1) {
		t.Error("after insert without contacts and groups ptrs must be equals")
	}

	inserted_group, err := ph.AddGroupToProfile(inserted_profile.UserName, &ProfileGroup{Name:"test_group", Description:"test_group_description"})
	check_err(t, err, "insert group")
	if inserted_group.Id == 0 {
		t.Error("after insert group must be have id")
	}

	inserted_contact, err := ph.AddContactToProfile(inserted_profile.UserName, &contact)
	check_err(t, err, "insert contact")
	if inserted_contact.ContactId == 0 {
		t.Error("after insert contact must have id")
	}

}
func TestProfilesUpdateGroups(t *testing.T) {
	t.SkipNow()
	ph, err := NewProfileDbHandler(config.Main.PGDatabase.ConnString)
	check_err(t, err, "init")
	deleteAll(ph)

	profile1PtrAfterInsert, _ := ph.InsertNewProfile(&profileWithGroup)
	if profile1PtrAfterInsert == nil {
		t.Error("After insert result  must be not nil")
	}
	check_groups(t, profile1PtrAfterInsert, 1)
	if !profile1PtrAfterInsert.Equal(&profileWithGroup) {
		t.Error("After insert profile and Before are not equals")
	}

	//add equal group and at result it must be one group
	profileWithGroup.Groups = append(profileWithGroup.Groups, group)
	err = ph.UpdateProfile(&profileWithGroup)
	check_err(t, err, "add group update")
	savedProfile2AfterUpdate, err := ph.GetProfile(profileWithGroup.UserName)
	check_err(t, err, "get updated group profile")
	check_groups(t, savedProfile2AfterUpdate, 1)

	//add another group
	profileWithGroup.Groups = append(profileWithGroup.Groups, ProfileGroup{Name:"test_Gr_2", Description:"foooooo"})
	err = ph.UpdateProfile(&profileWithGroup)
	check_err(t, err, "add group update")
	savedProfile3AfterUpdate, err := ph.GetProfile(profileWithGroup.UserName)
	check_err(t, err, "get updated group profile")
	check_groups(t, savedProfile3AfterUpdate, 2)
}

func TestProfilesUpdateContacts(t *testing.T) {
	t.SkipNow()
	ph, err := NewProfileDbHandler(config.Main.PGDatabase.ConnString)
	check_err(t, err, "init")
	deleteAll(ph)

	pwc, _ := ph.InsertNewProfile(&profileWithContacts)
	if pwc == nil {
		t.Error("After insert result  must be not nil")
	}
	if !check_contacts(t, pwc) {
		t.Error("Contacts or links not have id")
	}

	profileWithContacts.Contacts[0].Links[0].Description = "another_description"
	err = ph.UpdateProfile(&profileWithContacts)
	check_err(t, err, "update link description")
	saved, err := ph.GetProfile(profileWithContacts.UserName)
	check_err(t, err, "get saved update link description")
	if saved.Contacts[0].Links[0].Description != profileWithContacts.Contacts[0].Links[0].Description {
		t.Error("links description are not changed!", saved.Contacts[0].Links[0].Description, profileWithContacts.Contacts[0].Links[0].Description)
	}
	if len(saved.Contacts) != len(profileWithContacts.Contacts) {
		t.Error("added not added contact")
	}

	profileWithContacts.Contacts[0].Description = "another_description"
	err = ph.UpdateProfile(&profileWithContacts)
	check_err(t, err, "update description")
	saved, err = ph.GetProfile(profileWithContacts.UserName)
	check_err(t, err, "get saved update description")
	if saved.Contacts[0].Description != profileWithContacts.Contacts[0].Description {
		t.Error("description are not changed!", saved.Contacts[0].Description, profileWithContacts.Contacts[0].Description)
	}
	if len(saved.Contacts) != len(profileWithContacts.Contacts) {
		t.Error("added not added contact")
	}

	profileWithContacts.Contacts[0].Lat = 123.456
	profileWithContacts.Contacts[0].Lon = 456.123
	err = ph.UpdateProfile(&profileWithContacts)
	check_err(t, err, "update geo")
	saved, err = ph.GetProfile(profileWithContacts.UserName)
	check_err(t, err, "get saved update geo")
	if saved.Contacts[0].Lat != profileWithContacts.Contacts[0].Lat || saved.Contacts[0].Lon != profileWithContacts.Contacts[0].Lon {
		t.Error("geo are not changed!")
	}
	if len(saved.Contacts) != len(profileWithContacts.Contacts) {
		t.Error("added not added contact")
	}

	profileWithContacts.Contacts[0].OrderNumber = 100500
	err = ph.UpdateProfile(&profileWithContacts)
	check_err(t, err, "update OrderNumber")
	saved, err = ph.GetProfile(profileWithContacts.UserName)
	check_err(t, err, "get saved update OrderNumber")
	if saved.Contacts[0].OrderNumber != profileWithContacts.Contacts[0].OrderNumber {
		t.Error("OrderNumber are not changed!")
	}
	if len(saved.Contacts) != len(profileWithContacts.Contacts) {
		t.Error("added not added contact")
	}
}

func TestProfilesUpdateFields(t *testing.T) {
	t.SkipNow()
	ph, err := NewProfileDbHandler(config.Main.PGDatabase.ConnString)
	check_err(t, err, "init")
	deleteAll(ph)
	spe, _ := ph.InsertNewProfile(&profileEnabled)
	if !spe.Enable {
		t.Error("spe not enabled")
	}
	sppe, _ := ph.InsertNewProfile(&profilePublic)
	if !sppe.Public {
		t.Error("sppe not public")
	}
	spe.ShortDescription = "ttrtt"
	err = ph.UpdateProfile(spe)
	check_err(t, err, "short descr changed")
	supe, err := ph.GetProfile(spe.UserName)
	check_err(t, err, "short upd descr changed")
	if supe.ShortDescription != spe.ShortDescription {
		t.Error("short descr not changed")
	}

	sppe.TextDescription = "ttttttt"
	err = ph.UpdateProfile(sppe)
	check_err(t, err, "long descr changed")
	susppe, err := ph.GetProfile(sppe.UserName)
	check_err(t, err, "short upd descr changed")
	if susppe.TextDescription != sppe.TextDescription {
		t.Error("long descr not changed")
	}

	sppe.Public = false
	err = ph.UpdateProfile(sppe)
	check_err(t, err, "public changed")
	susppe, err = ph.GetProfile(sppe.UserName)
	check_err(t, err, "public upd changed")
	if susppe.Public != sppe.Public {
		t.Error("public not changed")
	}

	sppe.Enable = true
	err = ph.UpdateProfile(sppe)
	check_err(t, err, "Enable changed")
	susppe, err = ph.GetProfile(sppe.UserName)
	check_err(t, err, "Enable upd changed")
	if susppe.Enable != sppe.Enable {
		t.Error("Enable not changed")
	}

	sppe.ImageURL = "fooo"
	err = ph.UpdateProfile(sppe)
	check_err(t, err, "ImageURL changed")
	susppe, err = ph.GetProfile(sppe.UserName)
	check_err(t, err, "ImageURL upd changed")
	if susppe.ImageURL != sppe.ImageURL {
		t.Error("ImageURL not changed")
	}

	sppe.Name = "aaaaaaaaaa"
	err = ph.UpdateProfile(sppe)
	check_err(t, err, "Name changed")
	susppe, err = ph.GetProfile(sppe.UserName)
	check_err(t, err, "Name upd changed")
	if susppe.Name != sppe.Name {
		t.Error("Name not changed")
	}
}

func TestPhones(t *testing.T) {
	ph, err := NewProfileDbHandler(config.Main.PGDatabase.ConnString)
	check_err(t, err, "init")
	deleteAll(ph)
	phone_value := "79811064022"
	phone_value2 := "79138973664"
	p := &Profile{Name:"test phone", ShortDescription:"test phone", TextDescription:"test phone", UserName:"test_phone"}
	p.AllowedPhones = append(p.AllowedPhones, ProfileAllowedPhone{Value:phone_value})

	new_p, err := ph.InsertNewProfile(p)
	check_err(t, err, "insert profile with phone")
	if len(new_p.AllowedPhones) != 1{
		t.Errorf("Except one phone but %v",len(new_p.AllowedPhones))
	}
	if (new_p.AllowedPhones[0].Value != phone_value){
		t.Errorf("Except phone number %v but %v", phone_value, new_p.AllowedPhones[0].Value)
	}
	new_p.AllowedPhones = append(new_p.AllowedPhones, ProfileAllowedPhone{Value:phone_value2})
	ph.UpdateProfile(new_p)

	new_p, err = ph.GetProfile("test_phone")
	check_err(t,err, "get profile with two phones")

	if len(new_p.AllowedPhones) != 2{
		t.Errorf("Except two phones but %v",len(new_p.AllowedPhones))
	}
	if (new_p.AllowedPhones[1].Value != phone_value2){
		t.Errorf("Except phone number %v but %v", phone_value2, new_p.AllowedPhones[1].Value)
	}
}
