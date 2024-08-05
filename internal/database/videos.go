package database

import (
	log "github.com/sirupsen/logrus"
)

func GetAllVideosPerTopics() []Topic {
	var topics []Topic

	result := Instance.db.Model(&Topic{}).Preload("Subtopics.Videos").Find(&topics)

	if result.Error != nil {
		log.Error("Error fetching videos: ", result.Error)
	}

	return topics
}

func GetSingleVideo(videoId int) *Video {
	var video Video

	result := Instance.db.First(&video, videoId)

	if result.Error != nil {
		log.Error("Error fetching video: ", result.Error)
		return nil
	}

	return &video
}
