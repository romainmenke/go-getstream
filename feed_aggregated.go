package getstream

import (
	"encoding/json"
	"errors"
	"regexp"
	"strings"
)

// AggregatedFeed is a getstream AggregatedFeed
// Use it to for CRUD on AggregatedFeed Groups
type AggregatedFeed struct {
	Client   *Client
	feedSlug string
	userID   string
	token    string
}

// Signature is used to sign Requests : "FeedSlugUserID Token"
func (f *AggregatedFeed) Signature() string {
	if f.Token() == "" {
		return f.FeedIDWithoutColon()
	}
	return f.FeedIDWithoutColon() + " " + f.Token()
}

// FeedSlug returns the feed slug, it is needed to conform to the Feed interface
func (f *AggregatedFeed) FeedSlug() string {
	return f.feedSlug
}

// UserID returns the user id, it is needed to conform to the Feed interface
func (f *AggregatedFeed) UserID() string {
	return f.userID
}

// FeedIDWithColon is the combo of the FeedSlug and UserID : "FeedSlug:UserID"
func (f *AggregatedFeed) FeedIDWithColon() string {
	return f.FeedSlug() + ":" + f.UserID()
}

// FeedIDWithoutColon is the combo of the FeedSlug and UserID : "FeedSlugUserID"
func (f *AggregatedFeed) FeedIDWithoutColon() string {
	return f.FeedSlug() + f.UserID()
}

// SignFeed sets the token on a Feed
func (f *AggregatedFeed) SignFeed(signer *Signer) {
	if f.Client.Signer != nil {
		f.token = signer.GenerateToken(f.FeedIDWithoutColon())
	}
}

// Token returns the token of a Feed
func (f *AggregatedFeed) Token() string {
	return f.token
}

// GenerateToken returns a new Token for a Feed without setting it to the Feed
func (f *AggregatedFeed) GenerateToken(signer *Signer) string {
	if f.Client.Signer != nil {
		return signer.GenerateToken(f.FeedSlug() + f.UserID())
	}
	return ""
}

// AddActivity is used to add an Activity to a AggregatedFeed
func (f *AggregatedFeed) AddActivity(activity *Activity) (*Activity, error) {

	payload, err := json.Marshal(activity)
	if err != nil {
		return nil, err
	}

	endpoint := "feed/" + f.FeedSlug() + "/" + f.UserID() + "/"

	resultBytes, err := f.Client.post(f, endpoint, payload, nil)
	if err != nil {
		return nil, err
	}

	output := &Activity{}
	err = json.Unmarshal(resultBytes, output)
	if err != nil {
		return nil, err
	}

	return output, err
}

// AddActivities is used to add multiple Activities to a NotificationFeed
func (f *AggregatedFeed) AddActivities(activities []*Activity) ([]*Activity, error) {

	payload, err := json.Marshal(map[string][]*Activity{
		"activities": activities,
	})
	if err != nil {
		return nil, err
	}

	endpoint := "feed/" + f.FeedSlug() + "/" + f.UserID() + "/"

	resultBytes, err := f.Client.post(f, endpoint, payload, nil)
	if err != nil {
		return nil, err
	}

	output := &postAggregatedFeedOutputActivities{}
	err = json.Unmarshal(resultBytes, output)
	if err != nil {
		return nil, err
	}

	return output.Activities, err
}

// Activities returns a list of Activities for a NotificationFeedGroup
func (f *AggregatedFeed) Activities(input *GetActivitiesInput) (*GetAggregatedFeedOutput, error) {

	var payload []byte
	var err error

	if input != nil {
		payload, err = json.Marshal(input)
		if err != nil {
			return nil, err
		}
	}

	endpoint := "feed/" + f.FeedSlug() + "/" + f.UserID() + "/"

	result, err := f.Client.get(f, endpoint, payload, nil)
	if err != nil {
		return nil, err
	}

	output := &getAggregatedFeedOutput{}
	err = json.Unmarshal(result, output)
	if err != nil {
		return nil, err
	}

	return output.output(), err
}

// RemoveActivity removes an Activity from a NotificationFeedGroup
func (f *AggregatedFeed) RemoveActivity(input *Activity) error {

	endpoint := "feed/" + f.FeedSlug() + "/" + f.UserID() + "/" + input.ID + "/"

	return f.Client.del(f, endpoint, nil, nil)
}

// RemoveActivityByForeignID removes an Activity from a NotificationFeedGroup by ForeignID
func (f *AggregatedFeed) RemoveActivityByForeignID(input *Activity) error {

	if input.ForeignID == "" {
		return errors.New("no ForeignID")
	}

	r, err := regexp.Compile("^[a-z0-9]{8}-[a-z0-9]{4}-[1-5][a-z0-9]{3}-[a-z0-9]{4}-[a-z0-9]{12}$")
	if err != nil {
		return err
	}
	if !r.MatchString(input.ForeignID) {
		return errors.New("invalid ForeignID")
	}

	endpoint := "feed/" + f.FeedSlug() + "/" + f.UserID() + "/" + input.ForeignID + "/"

	return f.Client.del(f, endpoint, nil, map[string]string{
		"foreign_id": "1",
	})
}

// FollowFeedWithCopyLimit sets a Feed to follow another target Feed
// CopyLimit is the maximum number of Activities to Copy from History
func (f *AggregatedFeed) FollowFeedWithCopyLimit(target *FlatFeed, copyLimit int) error {
	endpoint := "feed/" + f.FeedSlug() + "/" + f.UserID() + "/" + "following" + "/"

	input := postAggregatedFeedFollowingInput{
		Target:            target.FeedIDWithColon(),
		ActivityCopyLimit: copyLimit,
	}

	payload, err := json.Marshal(input)
	if err != nil {
		return err
	}

	_, err = f.Client.post(f, endpoint, payload, nil)
	return err

}

// Unfollow is used to Unfollow a target Feed
func (f *AggregatedFeed) Unfollow(target *FlatFeed) error {

	endpoint := "feed/" + f.FeedSlug() + "/" + f.UserID() + "/" + "following" + "/" + target.FeedIDWithColon() + "/"

	return f.Client.del(f, endpoint, nil, nil)

}

// UnfollowKeepingHistory is used to Unfollow a target Feed while keeping the History
// this means that Activities already visibile will remain
func (f *AggregatedFeed) UnfollowKeepingHistory(target *FlatFeed) error {

	endpoint := "feed/" + f.FeedSlug() + "/" + f.UserID() + "/" + "following" + "/" + target.FeedIDWithColon() + "/"

	payload, err := json.Marshal(map[string]string{
		"keep_history": "1",
	})
	if err != nil {
		return err
	}

	return f.Client.del(f, endpoint, payload, nil)

}

// FollowersWithLimitAndSkip returns a list of GeneralFeed following the current AggregatedFeed
func (f *AggregatedFeed) FollowersWithLimitAndSkip(limit int, skip int) ([]*GeneralFeed, error) {
	var err error

	endpoint := "feed/" + f.FeedSlug() + "/" + f.UserID() + "/" + "followers" + "/"

	payload, err := json.Marshal(&getFlatFeedFollowersInput{
		Limit: limit,
		Skip:  skip,
	})
	if err != nil {
		return nil, err
	}

	resultBytes, err := f.Client.get(f, endpoint, payload, nil)

	output := &getAggregatedFeedFollowersOutput{}
	err = json.Unmarshal(resultBytes, output)
	if err != nil {
		return nil, err
	}

	var outputFeeds []*GeneralFeed
	for _, result := range output.Results {

		feed := GeneralFeed{}

		var match bool
		match, err = regexp.MatchString(`^.*?:.*?$`, result.FeedID)
		if err != nil {
			continue
		}

		if match {
			firstSplit := strings.Split(result.FeedID, ":")

			feed.feedSlug = firstSplit[0]
			feed.userID = firstSplit[1]
		}

		outputFeeds = append(outputFeeds, &feed)
	}

	return outputFeeds, err
}

// FollowingWithLimitAndSkip returns a list of GeneralFeed followed by the current FlatFeed
func (f *AggregatedFeed) FollowingWithLimitAndSkip(limit int, skip int) ([]*GeneralFeed, error) {

	var err error

	endpoint := "feed/" + f.FeedSlug() + "/" + f.UserID() + "/" + "following" + "/"

	payload, err := json.Marshal(&getAggregatedFeedFollowersInput{
		Limit: limit,
		Skip:  skip,
	})
	if err != nil {
		return nil, err
	}

	resultBytes, err := f.Client.get(f, endpoint, payload, nil)

	output := &getAggregatedFeedFollowersOutput{}
	err = json.Unmarshal(resultBytes, output)
	if err != nil {
		return nil, err
	}

	var outputFeeds []*GeneralFeed
	for _, result := range output.Results {

		feed := GeneralFeed{}

		var match bool
		match, err = regexp.MatchString(`^.*?:.*?$`, result.FeedID)
		if err != nil {
			continue
		}

		if match {
			firstSplit := strings.Split(result.TargetID, ":")

			feed.feedSlug = firstSplit[0]
			feed.userID = firstSplit[1]
		}

		outputFeeds = append(outputFeeds, &feed)
	}

	return outputFeeds, err

}
