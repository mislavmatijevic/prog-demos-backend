package database

import "time"

type Topic struct {
	ID        int         `gorm:"primaryKey" json:"id"`
	Name      string      `gorm:"size:100;not null" json:"name"`
	Subtopics []*Subtopic `gorm:"foreignKey:TopicID" json:"subtopics"`
}

type Subtopic struct {
	ID      int          `gorm:"primaryKey" json:"id"`
	TopicID int          `gorm:"not null;column:id_topic" json:"-"`
	Name    string       `gorm:"size:100;not null" json:"name"`
	Topic   *Topic       `gorm:"foreignKey:TopicID" json:"topic,omitempty"`
	Videos  []*Video     `gorm:"foreignKey:SubtopicID" json:"videos,omitempty"`
	Tasks   []*BasicTask `gorm:"foreignKey:SubtopicID" json:"tasks,omitempty"`
}

type Video struct {
	ID         int       `gorm:"primaryKey" json:"id"`
	SubtopicID int       `gorm:"not null;column:id_subtopic" json:"-"`
	Identifier string    `gorm:"type:char(11);not null" json:"identifier"`
	Name       string    `gorm:"size:255;not null" json:"name"`
	Subtopic   *Subtopic `gorm:"foreignKey:SubtopicID" json:"subtopic,omitempty"`
}

type BasicTask struct {
	ID         int `gorm:"primaryKey" json:"id"`
	SubtopicID int `gorm:"not null;column:id_subtopic" json:"-"`
	OrderNum   int `gorm:"not null" json:"order_num"`
}

func (BasicTask) TableName() string {
	return "tasks"
}

type FullTask struct {
	ID                 int       `gorm:"primaryKey" json:"id"`
	SubtopicID         int       `gorm:"not null;column:id_subtopic" json:"id_subtopic"`
	OrderNum           int       `gorm:"not null" json:"order_num"`
	Input              string    `gorm:"size:512;not null" json:"input"`
	Output             string    `gorm:"size:512;not null" json:"output"`
	InputOutputExample string    `gorm:"type:text" json:"input_output_example"`
	IsFinalBoss        bool      `gorm:"not null;default:false" json:"is_final_boss"`
	StarterCode        string    `gorm:"type:text" json:"starter_code"`
	Step1Code          string    `gorm:"type:text" json:"step1_code,omitempty"`
	Step2Code          string    `gorm:"type:text" json:"step2_code,omitempty"`
	Step3Code          string    `gorm:"type:text" json:"step3_code,omitempty"`
	Helper1Text        string    `gorm:"size:255" json:"helper1_text,omitempty"`
	Helper2Text        string    `gorm:"size:255" json:"helper2_text,omitempty"`
	Helper3Text        string    `gorm:"size:255" json:"helper3_text,omitempty"`
	SolutionCode       string    `gorm:"type:text" json:"solution_code"`
	Subtopic           *Subtopic `gorm:"foreignKey:SubtopicID" json:"subtopic"`
	Tests              []Test    `gorm:"foreignKey:IDTask" json:"-"`
}

func (FullTask) TableName() string {
	return "tasks"
}

type Test struct {
	ID             int      `gorm:"primaryKey" json:"id"`
	IDTask         int      `gorm:"not null" json:"id_task"`
	Input          string   `gorm:"size:512;not null" json:"input"`
	ExpectedOutput string   `gorm:"size:512;not null" json:"expected_output"`
	Task           FullTask `gorm:"foreignKey:IDTask" json:"task"`
	ArtefactSHA256 string   `gorm:"type:char(64)" json:"-"`
}

type User struct {
	ID              int       `gorm:"primaryKey" json:"id"`
	Username        string    `gorm:"not null" json:"username"`
	Email           string    `gorm:"not null" json:"email"`
	Password        string    `gorm:"not null" json:"-"`
	UserType        string    `gorm:"not null;default:basic" json:"-"`
	IsActivated     bool      `gorm:"not null;default:false" json:"-"`
	ActivationToken string    `gorm:"type:char(128);not null;default:false" json:"-"`
	DateRegistered  time.Time `gorm:"not null" json:"-"`
}

type RefreshToken struct {
	ID         int       `gorm:"primaryKey" json:"-"`
	OwnerID    int       `gorm:"not null;column:id_user" json:"-"`
	Value      string    `gorm:"type:char(512)" json:"refresh_token"`
	Expiration time.Time `gorm:"not null" json:"expires_at"`
	Owner      *User     `gorm:"foreignKey:OwnerID" json:"-"`
}
