package voting

import (
	"testing"
	"msngr/configuration"
	"log"
	"gopkg.in/mgo.v2/bson"
	"msngr/test"
	"msngr/utils"
)

var config = configuration.ReadTestConfigInRecursive()

func clear(vdh *VotingDataHandler) {
	vdh.Companies.RemoveAll(bson.M{})
}

func PrepObj() *VotingDataHandler {
	vdh, err := NewVotingHandler(config.Main.Database.ConnString, config.Main.Database.Name)
	if err != nil {
		log.Panic(err)
	}
	clear(vdh)
	return vdh
}

func TestVoteCompany(t *testing.T) {
	vdh := PrepObj()
	err := vdh.ConsiderCompany("test company", "NSK", "test", "", "test_user", "test_role")
	if err != nil {
		t.Errorf("first add cmp %v", err)
	}
	err = vdh.ConsiderCompany("test company", "NSK", "test", "", "test_user1", "test_role")
	if err != nil {
		t.Errorf("err at add same cmp but another user %v", err)
	}

	err = vdh.ConsiderCompany("test company", "NSK", "test", "", "test_user", "test_role1")
	if err == nil {
		t.Errorf("err is none when add same cmp with same username but another user role %v", err)
	}
	companies, err := vdh.GetCompanies(bson.M{"name":"test company"})
	if err != nil {
		t.Errorf("err at getting cmps %v", err)
	}
	test.CheckCount(companies, 1, t, "companies count")
	test.CheckCount(companies[0].VoteInfo.Voters, 2, t, "voters at company")
	test.CheckEquals(companies[0].Get("name"), "test company", t, "name field must be equals")
	test.CheckEquals(companies[0].Get("city"), "NSK", t, "name field must be equals")
	test.CheckEquals(companies[0].Get("service"), "test", t, "name field must be equals")

	vdh.ConsiderCompany("test company 2", "NSK", "", "", "test_user", "")
	all_companies, err := vdh.GetCompanies(bson.M{})
	if err != nil {
		t.Errorf("err at getting cmps %v", err)
	}
	test.CheckCount(all_companies, 2, t, "companies count after add another")

	vdh.ConsiderCompany("test company 2", "MSK", "", "", "test_user", "")
	all_companies, err = vdh.GetCompanies(bson.M{})
	if err != nil {
		t.Errorf("err at getting cmps %v", err)
	}
	test.CheckCount(all_companies, 3, t, "companies count after add another")
}

func TestAutocomplete(t *testing.T) {
	vdh := PrepObj()
	vdh.ConsiderCompany("abc", "NSK", "abc", "", "", "")
	vdh.ConsiderCompany("aabc", "NSK_", "abc cba", "", "", "")
	vdh.ConsiderCompany("aabbc", "NS_K", "abc qwe", "", "", "")
	vdh.ConsiderCompany("aabbcc", "_NSK", "qwe abc ", "", "", "")

	res, err := vdh.TextFoundByCompanyField("a", "name")
	test.CheckErr(t, err, "found by company field")
	test.CheckCount(res, 4, t, "by 'a' and 'name' must be all")

	res, err = vdh.TextFoundByCompanyField("ab", "name")
	test.CheckErr(t, err, "found by company field")
	test.CheckCount(res, 4, t, "by 'ab' and 'name' must be all")

	res, err = vdh.TextFoundByCompanyField("ab", "service")
	test.CheckErr(t, err, "found by company field")
	test.CheckCount(res, 4, t, "by 'ab' and 'service' must be all")

	res, err = vdh.TextFoundByCompanyField("cc", "name")
	test.CheckErr(t, err, "found by company field")
	test.CheckCount(res, 1, t, "by 'c' and 'name' must be all")

	res, err = vdh.TextFoundByCompanyField("bb", "name")
	test.CheckErr(t, err, "found by company field")
	test.CheckCount(res, 2, t, "by 'bb' and 'name' must be all")
	if !utils.InS("aabbc", res) {
		t.Errorf("interested aabbc not in result: %+v", res)
	}

	res, err = vdh.TextFoundByCompanyField("_", "city")
	test.CheckErr(t, err, "found by company field")
	test.CheckCount(res, 3, t, "by '_' and 'city' must be all")
	if !utils.InS("NS_K", res) || !utils.InS("NSK_", res) || !utils.InS("_NSK", res) {
		t.Errorf("interested names not in result: %+v", res)
	}
}

func TestUserVotes(t *testing.T) {
	vdh := PrepObj()
	vdh.ConsiderCompany("abc", "NSK", "abc", "foo", "u1", "")
	vdh.ConsiderCompany("abc1", "NSK", "abc", "foo", "u1", "")
	vdh.ConsiderCompany("abc2", "NSK", "abc", "foo", "u1", "")

	vdh.ConsiderCompany("abc2", "NSK", "abc", "foo", "u2", "")
	vdh.ConsiderCompany("abc3", "NSK", "abc", "foo", "u2", "")
	vdh.ConsiderCompany("abc4", "NSK", "abc", "foo", "u2", "")
	vdh.ConsiderCompany("abc5", "NSK", "abc", "foo", "u2", "")

	cmps, err := vdh.GetUserVotes("u1")
	test.CheckErr(t, err, "get votes error")
	test.CheckCount(cmps, 3, t, "for u1 must be 3 votes")

	vdh.ConsiderCompany("abc2", "NSK", "abc", "foo", "u2", "")
	vdh.ConsiderCompany("abc3", "NSK", "abc", "foo", "u2", "")
	vdh.ConsiderCompany("abc4", "NSK", "abc", "foo", "u2", "")
	vdh.ConsiderCompany("abc5", "NSK", "abc", "foo", "u2", "")

	cmps, err = vdh.GetUserVotes("u2")
	test.CheckErr(t, err, "get votes error")
	test.CheckCount(cmps, 4, t, "for u2 must be 4 votes")

}