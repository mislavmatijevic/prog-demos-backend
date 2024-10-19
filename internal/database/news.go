package database

func GetNews() (news []News) {
	Instance.db.Model(&News{}).Find(&news)
	return
}
