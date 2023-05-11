package models

type Customer struct {
	ID          string `json:"id,omitempty" bson:"_id,omitempty"`
	Name        string `json:"name,omitempty" bson:"name,omitempty"`
	Password    string `json:"password,omitempty" bson:"password,omitempty"`
	Email       string `json:"email,omitempty" bson:"email,omitempty"`
	Phone       string `json:"phone,omitempty" bson:"phone,omitempty"`
	AffiliateID string `json:"affiliate_id,omitempty" bson:"affiliate_id,omitempty"`
}
