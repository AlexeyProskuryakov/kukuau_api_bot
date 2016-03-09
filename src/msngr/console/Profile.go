package console

type ProfileContact struct {
	Id          int64 `json:"id"`
	Type        string `json:"type"`
	Value       string `json:"value"`
	Description string `json:"description"`
	ShowedText  string `json:"showed_text"`
}

type Coordinates struct {
	Lat float64 `json:"lat"`
	Lon float64 `json:"lon"`
}

type Profile struct {
	Id               int64 `json:"id"`
	ImageURL         string `json:"image_url"`
	Name             string `json:"name"`
	ShortDescription string `json:"short_description"`
	TextDescription  string `json:"text_description"`
	Contacts         []ProfileContact `json:"contacts"`
	Address          string `json:"address"`
	Place            Coordinates `json:"place"`
}

func GetAllProfiles() []Profile {
	return []Profile{}
}

func CreateProfile() {

}

func UpdateProfile() {

}

func DeleteProfile() {

}

