package database

import (
	"database/sql"
	"time"
)

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
	ID         int    `gorm:"primaryKey" json:"id"`
	Name       string `gorm:"not null;column:name" json:"name"`
	SubtopicID int    `gorm:"not null;column:id_subtopic" json:"-"`
	Complexity int    `gorm:"not null" json:"complexity"`
}

func (BasicTask) TableName() string {
	return "tasks"
}

type FullTask struct {
	ID                 int            `gorm:"primaryKey" json:"id"`
	Name               string         `gorm:"not null;column:name" json:"name"`
	SubtopicID         int            `gorm:"not null;column:id_subtopic" json:"idSubtopic"`
	CreatorID          int            `gorm:"not null;column:id_user" json:"-"`
	Complexity         string         `gorm:"type:char(1);not null" json:"complexity"`
	Input              string         `gorm:"size:512;not null" json:"input"`
	Output             string         `gorm:"size:512;not null" json:"output"`
	InputOutputExample string         `gorm:"type:text" json:"inputOutputExample"`
	IsFinalBoss        bool           `gorm:"not null;default:false" json:"isFinalBoss"`
	SolutionCode       string         `gorm:"type:text" json:"solutionCode,omitempty"`
	Subtopic           *Subtopic      `gorm:"foreignKey:SubtopicID" json:"subtopic"`
	Tests              []TaskTest     `gorm:"foreignKey:IDTask" json:"-"`
	Creator            *User          `gorm:"foreignKey:CreatorID" json:"creator,omitempty"`
	HelpSteps          []TaskHelpStep `gorm:"foreignKey:IDTask" json:"helpSteps"`
}

func (FullTask) TableName() string {
	return "tasks"
}

type TaskHelpStep struct {
	ID         int            `gorm:"primaryKey" json:"-"`
	IDTask     int            `gorm:"not null" json:"idTask"`
	Step       int            `gorm:"int" json:"step"`
	HelperCode sql.NullString `gorm:"type:text" json:"helperCode,omitempty"`
	HelperText sql.NullString `gorm:"size:512" json:"helperText,omitempty"`
}

type TaskTest struct {
	ID             int    `gorm:"primaryKey" json:"id"`
	IDTask         int    `gorm:"not null" json:"idTask"`
	Input          string `gorm:"size:512;not null" json:"input"`
	ExpectedOutput string `gorm:"size:512;not null" json:"expectedOutput"`
	ArtefactSHA256 string `gorm:"type:char(64)" json:"-"`
}

type User struct {
	ID                  int            `gorm:"primaryKey" json:"id"`
	Username            string         `gorm:"not null" json:"username"`
	Email               string         `gorm:"not null" json:"email"`
	Password            string         `gorm:"not null" json:"-"`
	UserType            string         `gorm:"not null;default:basic" json:"type"`
	IsActivated         bool           `gorm:"not null;default:false" json:"-"`
	ActivationToken     sql.NullString `gorm:"type:char(128);null" json:"-"`
	DateRegistered      time.Time      `gorm:"not null" json:"-"`
	PasswordResetToken  string         `gorm:"type:char(128);null;default:null" json:"-"`
	PasswordResetExpiry time.Time      `gorm:"null;default:null" json:"-"`
}

type RefreshToken struct {
	ID         int       `gorm:"primaryKey" json:"-"`
	OwnerID    int       `gorm:"not null;column:id_user" json:"-"`
	Value      string    `gorm:"type:char(512)" json:"value"`
	Expiration time.Time `gorm:"not null" json:"expiresAt"`
	Owner      *User     `gorm:"foreignKey:OwnerID" json:"-"`
}

type TaskExecution struct {
	ID            int       `gorm:"primaryKey" json:"-"`
	InitiatorID   int       `gorm:"not null;column:id_user" json:"-"`
	TaskID        int       `gorm:"not null;column:id_task" json:"-"`
	IsFinished    bool      `gorm:"not null;default:false" json:"-"`
	StartedAt     time.Time `gorm:"not null" json:"-"`
	FinishedAt    time.Time `gorm:"null;default:null" json:"-"`
	SubmittedCode string    `gorm:"type:text" json:"solutionCode"`
	WasSuccessful bool      `gorm:"not null;default:false" json:"-"`
	Initiator     *User     `gorm:"foreignKey:InitiatorID" json:"-"`
	Task          *FullTask `gorm:"foreignKey:TaskID" json:"-"`
}
