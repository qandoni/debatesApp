package core_realtime

import "strconv"

const TopicPostPrefix = "post:"

func PostTopic(postID int) string {
	return TopicPostPrefix + strconv.Itoa(postID)
}
