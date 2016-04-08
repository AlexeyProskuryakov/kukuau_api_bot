package voting

import (
	d "msngr/db"
	"gopkg.in/mgo.v2"
	"errors"
	"gopkg.in/mgo.v2/bson"
	"fmt"
	"msngr/utils"
	"reflect"
	"time"
)

type Voter struct {
	UserName string `bson:"user_name"`
	Role     string `bson:"role,omitempty"`
	VoteTime time.Time `bson:"vote_time"`
}

func (v Voter) String() string {
	if v.Role != "" {
		return fmt.Sprintf("%v (%v)", v.UserName, v.Role)
	}
	return v.UserName
}

type VoteObject struct {
	VoteCount int `bson:"vote_count"`
	Voters    []Voter `bson:"voters"`
}

func (vo VoteObject) String() string {
	return fmt.Sprintf("\n\tcount: %v, users:%+v", vo.VoteCount, vo.Voters)
}
func (vo VoteObject) ContainUserName(userName string) bool {
	for _, fVouter := range vo.Voters {
		if fVouter.UserName == userName {
			return true
		}
	}
	return false
}

type CompanyModel struct {
	VoteInfo    VoteObject `bson:"vote"`
	ID          bson.ObjectId `bson:"_id,omitempty"`
	Name        string `bson:"name"`
	City        string `bson:"city"`
	Service     string `bson:"service"`
	Description string `bson:"description"`
}

func (cm CompanyModel) GetFieldValue(fieldBsonName string) string {
	v := reflect.ValueOf(cm)
	typ := v.Type()
	for i := 0; i < v.NumField(); i++ {
		// gets us a StructField
		fi := typ.Field(i)
		if tagv := fi.Tag.Get("bson"); tagv == fieldBsonName {
			return v.Field(i).String()
		}
	}
	return ""
}

func (cm CompanyModel) GetVoter(userName string) *Voter {
	for _, voter := range cm.VoteInfo.Voters {
		if voter.UserName == userName {
			return &voter
		}
	}
	return nil
}

type UserCompaniesMapping struct {
	UserName         string `bson:"user_name"`
	LastCompanyAdded bson.ObjectId `bson:"last_company_added"`
}

func (cm CompanyModel) ToMap() map[string]string {
	result := map[string]string{}
	result["name"] = cm.Name
	result["city"] = cm.City
	result["service"] = cm.Service
	result["description"] = cm.Description
	result["vote_count"] = fmt.Sprintf("%v человек", cm.VoteInfo.VoteCount)
	return result
}
func (cm CompanyModel) String() string {
	return fmt.Sprintf("\n-------------------\nCompany: [%v] \nName:%v\nCity:%v\nDescription:%v\nVotes:%+v\n-------------------\n",
		cm.ID, cm.Name, cm.City, cm.Description, cm.VoteInfo)
}

type VotingDataHandler struct {
	d.DbHelper
	Companies *mgo.Collection
}

func (vdh *VotingDataHandler) ensureIndexes() {
	companiesCollection := vdh.Session.DB(vdh.DbName).C("vote_companies")
	companiesCollection.EnsureIndex(mgo.Index{
		Key:        []string{"name", "city", "service"},
		Background: true,
		DropDups:   true,
		Unique:    true,
	})
	companiesCollection.EnsureIndex(mgo.Index{
		Key:        []string{"vote.voters.user_name"},
		Background: true,
		Unique:    false,
	})
	companiesCollection.EnsureIndex(mgo.Index{
		Key:        []string{"vote.voters.role"},
		Background: true,
		Unique:    false,
	})
	companiesCollection.EnsureIndex(mgo.Index{
		Key:        []string{"vote.voters.vote_time"},
		Background: true,
		Unique:    false,
	})
	companiesCollection.EnsureIndex(mgo.Index{
		Key:        []string{"vote.vote_count"},
		Background: true,
		Unique:    false,
	})
	companiesCollection.EnsureIndex(mgo.Index{
		Key:        []string{"name"},
		Background: true,
		Unique:    false,
	})
	companiesCollection.EnsureIndex(mgo.Index{
		Key:        []string{"city"},
		Background: true,
		Unique:    false,
	})
	companiesCollection.EnsureIndex(mgo.Index{
		Key:        []string{"service"},
		Background: true,
		Unique:    false,
	})
	vdh.Companies = companiesCollection
}

func NewVotingHandler(conn, dbName string) (*VotingDataHandler, error) {
	dbh := d.NewDbHelper(conn, dbName)
	if dbh.Check() {
		result := VotingDataHandler{DbHelper:*dbh}
		result.ensureIndexes()
		return &result, nil
	}
	return nil, errors.New("Can not connect to db, try it next time")
}

type AlreadyConsider struct {
	S string
}

func (ac AlreadyConsider) Error() string {
	return ac.S
}

func (vdh *VotingDataHandler) ConsiderCompany(name, city, service, description, userName, userRole string) (*CompanyModel, error) {
	found := CompanyModel{}
	err := vdh.Companies.Find(bson.M{"name":name, "city":city, "service":service}).One(&found)
	if err == mgo.ErrNotFound {
		toInsert := CompanyModel{
			Name:name,
			City:city,
			Description:description,
			Service:service,
			VoteInfo:VoteObject{
				Voters:[]Voter{
					Voter{UserName:userName, Role: userRole, VoteTime:time.Now()}},
				VoteCount:1,
			},
		}
		err = vdh.Companies.Insert(&toInsert)
		return &toInsert, err
	} else if err == nil {
		if found.VoteInfo.ContainUserName(userName) {
			return nil, AlreadyConsider{S:"Пользователь уже добавил эту компанию"}
		}
		voter := Voter{UserName:userName, Role:userRole}
		err = vdh.Companies.UpdateId(found.ID, bson.M{
			"$inc":bson.M{"vote.vote_count": 1},
			"$addToSet":bson.M{"vote.voters":voter},
		})
	} else {
		return nil, err
	}
	return &found, err
}

func (vdh *VotingDataHandler) GetCompanies(q bson.M) ([]CompanyModel, error) {
	result := []CompanyModel{}
	err := vdh.Companies.Find(q).All(&result)
	return result, err
}

func (vdh *VotingDataHandler) GetLastVote(userName string) (*CompanyModel, error) {
	result := CompanyModel{}
	err := vdh.Companies.Find(bson.M{"vote.voters.user_name":userName}).Sort("-vote.voters.vote_time").One(&result)
	if err == mgo.ErrNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (vdh *VotingDataHandler) TextFoundByCompanyField(q, field string) ([]string, error) {
	result := []string{}
	qResult := []CompanyModel{}
	if !utils.InS(field, []string{"name", "city", "service", "vote.voters.role"}) {
		return result, errors.New("field invalid")
	}
	err := vdh.Companies.Find(bson.M{field:bson.RegEx{fmt.Sprintf(".*%v.*", q), ""}}).All(&qResult)
	if err != nil && err != mgo.ErrNotFound {
		return result, err
	} else if err == mgo.ErrNotFound {
		return result, nil
	}
	for _, cm := range qResult {
		result = append(result, cm.GetFieldValue(field))
	}
	return result, nil
}

func (vdh *VotingDataHandler) GetUserVotes(username string) ([]CompanyModel, error) {
	result := []CompanyModel{}
	err := vdh.Companies.Find(bson.M{"vote.voters.user_name":username}).Sort("-vote.voters.vote_time").All(&result)
	return result, err
}

func (vdh *VotingDataHandler) GetTopVotes(limit int) ([]CompanyModel, error) {
	result := []CompanyModel{}
	q := vdh.Companies.Find(bson.M{}).Sort("-vote.vote_count")
	if limit > 0 {
		q.Limit(limit)
	}
	err := q.All(&result)
	return result, err
}