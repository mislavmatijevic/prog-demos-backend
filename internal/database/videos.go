package database

import (
	log "github.com/sirupsen/logrus"
)

func GetAllVideosPerTopics() []Topic {
	var topics []Topic

	result := Instance.db.Model(&[]Topic{}).Preload("Subtopics.Videos").Find(&topics)

	if result.Error != nil {
		log.WithError(result.Error).WithFields(log.Fields{"priority": "high", "context": "videos"}).Error("Couldn't fetch videos per topics!")
	}

	return topics
}

func GetSingleVideo(videoId int) *Video {
	var video Video

	result := Instance.db.First(&video, videoId)

	if result.Error != nil {
		return nil
	}

	return &video
}
