package simple

type MongoStorePageHtml struct {
	Id             string `bson:"_id" json:"id"`
	WebsiteName    string `bson:"websiteName" json:"websiteName"`
	Url            string `bson:"url"`
	Content        string `bson:"content"`
	HttpStatusCode byte   `bson:"httpStatusCode"`
	ContentSize    int    `bson:"contentSize"`
}
