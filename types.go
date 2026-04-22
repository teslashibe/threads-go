package threads

import (
	"encoding/json"
	"time"
)

// Cookies holds the Threads web session cookies obtained from a logged-in
// browser at threads.com. All five fields are required for cookie auth.
//
// SessionID is the primary credential. CSRFToken is also sent as the
// X-CSRFToken header on every request.
type Cookies struct {
	SessionID string `json:"sessionid"`  // sessionid: primary auth credential (URL-encoded)
	CSRFToken string `json:"csrftoken"`  // csrftoken: also sent as X-CSRFToken header
	DSUserID  string `json:"ds_user_id"` // ds_user_id: numeric user ID of logged-in account
	Mid       string `json:"mid"`        // mid: machine ID
	IgDid     string `json:"ig_did"`     // ig_did: device ID
}

// Auth holds the Bearer token credentials required for write operations.
// These are obtained via the Bloks login flow against i.instagram.com.
//
// Token is the IGT:2:... bearer string returned by the login response.
// UserID is the numeric ID of the authenticated user. DeviceID is the
// stable android-{13chars} identifier used during login and signing.
type Auth struct {
	Token    string `json:"token"`     // IGT:2:... bearer token from Bloks login
	UserID   string `json:"user_id"`   // numeric user ID
	DeviceID string `json:"device_id"` // android-{13chars} — stable per device
}

// User represents a Threads / Instagram user profile. Fields populated
// vary by source endpoint (search vs. profile vs. /accounts/current_user).
type User struct {
	ID                       string   `json:"id"`                            // pk / strong_id__: numeric user ID
	FBID                     string   `json:"fbid,omitempty"`                // fbid_v2: Meta-level user ID
	Username                 string   `json:"username"`                      // @handle (no @ prefix)
	FullName                 string   `json:"fullName"`                      // display name
	Biography                string   `json:"biography,omitempty"`           // text_app_biography concatenation
	IsPrivate                bool     `json:"isPrivate"`                     // private profile
	IsVerified               bool     `json:"isVerified"`                    // verified badge
	ProfilePicURL            string   `json:"profilePicUrl,omitempty"`       // profile_pic_url
	HDProfilePicURL          string   `json:"hdProfilePicUrl,omitempty"`     // hd_profile_pic_url_info.url
	HDProfilePicVersions     []Image  `json:"hdProfilePicVersions,omitempty"` // multiple sizes
	FollowerCount            int      `json:"followerCount"`                 // follower_count
	FollowingCount           int      `json:"followingCount"`                // following_count
	MediaCount               int      `json:"mediaCount,omitempty"`          // total media posts
	HasNMEBadge              bool     `json:"hasNmeBadge,omitempty"`
	ShowFBLinkOnProfile      bool     `json:"showFbLinkOnProfile,omitempty"`
	ThirdPartyDownloads      int      `json:"thirdPartyDownloadsEnabled,omitempty"`
	TextAppLowLike           bool     `json:"textAppIsLowLike,omitempty"`
	FeedPostReshareDisabled  bool     `json:"feedPostReshareDisabled,omitempty"`
	EligibleForActivBadge    bool     `json:"eligibleForTextAppActivationBadge,omitempty"`
	HideActivBadgeOnTextApp  bool     `json:"hideTextAppActivationBadgeOnTextApp,omitempty"`
	BiographyLinks           []string `json:"biographyLinks,omitempty"`
	ExternalURL              string   `json:"externalUrl,omitempty"`
}

// Image is a single sized profile picture / media variant.
type Image struct {
	URL    string `json:"url"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
}

// Post represents a single Threads post (alias: thread item / media).
type Post struct {
	ID                 string    `json:"id"`                            // pk / strong_id__: numeric post ID
	FBID               string    `json:"fbid,omitempty"`                // Meta-level post ID
	Code               string    `json:"code"`                          // base64 short code, used in URL
	TakenAt            time.Time `json:"takenAt"`                       // post creation timestamp
	MediaType          int       `json:"mediaType"`                     // 19=text, 1=photo, 2=video, 8=carousel
	ProductType        string    `json:"productType,omitempty"`         // typically "text_post"
	LikeCount          int       `json:"likeCount"`
	ReplyCount         int       `json:"replyCount,omitempty"`          // text_post_app_info.direct_reply_count
	RepostCount        int       `json:"repostCount,omitempty"`         // text_post_app_info.repost_count
	QuoteCount         int       `json:"quoteCount,omitempty"`          // text_post_app_info.quote_count
	Text               string    `json:"text"`                          // caption.text
	User               User      `json:"user"`                          // author (partial)
	HasLiked           bool      `json:"hasLiked"`                      // viewer's own like
	CanReshare         bool      `json:"canReshare,omitempty"`
	CanReply           bool      `json:"canReply,omitempty"`
	IsReplyToAuthor    bool      `json:"isReplyToAuthor,omitempty"`     // is_reply_to_author of root
	Permalink          string    `json:"permalink,omitempty"`           // canonical https://threads.com/...
	OriginalWidth      int       `json:"originalWidth,omitempty"`
	OriginalHeight     int       `json:"originalHeight,omitempty"`
	ImageVersions      []Image   `json:"imageVersions,omitempty"`       // image_versions2.candidates
	VideoVersions      []Video   `json:"videoVersions,omitempty"`       // video_versions
	CarouselMedia      []Post    `json:"carouselMedia,omitempty"`       // for media_type 8
	QuotedPost         *Post     `json:"quotedPost,omitempty"`          // text_post_app_info.share_info.quoted_post
	RepostedPost       *Post     `json:"repostedPost,omitempty"`        // text_post_app_info.share_info.reposted_post
	LinkPreview        *LinkPreview `json:"linkPreview,omitempty"`      // text_post_app_info.link_preview_attachment
	CaptionEdited      bool      `json:"captionEdited,omitempty"`
	OrganicTracking    string    `json:"organicTrackingToken,omitempty"`
}

// Video is a single video rendition.
type Video struct {
	URL    string `json:"url"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
	Type   int    `json:"type"`
}

// LinkPreview is a link card attached to a post.
type LinkPreview struct {
	URL         string `json:"url"`
	DisplayURL  string `json:"displayUrl,omitempty"`
	Title       string `json:"title,omitempty"`
	ImageURL    string `json:"imageUrl,omitempty"`
}

// Thread groups one or more posts together (a chain of thread items).
// For a single standalone post, ThreadItems will contain exactly one element.
type Thread struct {
	ID          string `json:"id"`
	ThreadItems []Post `json:"threadItems"`
}

// PostPage is one page of posts from a feed/listing endpoint, with a
// next_max_id cursor. NextCursor is empty when no further pages exist.
type PostPage struct {
	Threads    []Thread `json:"threads"`
	NextCursor string   `json:"nextCursor,omitempty"`
	HasNext    bool     `json:"hasNext"`
}

// UserPage is one page of users from a search/list endpoint.
type UserPage struct {
	Users      []User `json:"users"`
	NextCursor string `json:"nextCursor,omitempty"`
	HasNext    bool   `json:"hasNext"`
	NumResults int    `json:"numResults,omitempty"`
}

// ThreadContext is the response from /text_feed/{id}/replies/.
// It contains the focal post (in ContainingThread), top-level replies,
// sibling threads, and a downward cursor for paginating replies.
type ThreadContext struct {
	TargetPostID                 string   `json:"targetPostId"`
	ContainingThread             Thread   `json:"containingThread"`
	ReplyThreads                 []Thread `json:"replyThreads"`
	SiblingThreads               []Thread `json:"siblingThreads,omitempty"`
	DownwardCursor               string   `json:"downwardCursor,omitempty"`
	DownwardsThreadWillContinue  bool     `json:"downwardsThreadWillContinue"`
	IsSubscribedToTargetPost     bool     `json:"isSubscribedToTargetPost"`
	IsAuthorOfRootPost           bool     `json:"isAuthorOfRootPost"`
}

// FriendshipStatus describes the relationship between the viewer and a
// target user, as returned by /friendships/show/{id}/.
type FriendshipStatus struct {
	UserID                  string `json:"userId,omitempty"`
	Following               bool   `json:"following"`
	FollowedBy              bool   `json:"followedBy"`
	Blocking                bool   `json:"blocking"`
	Muting                  bool   `json:"muting"`
	IsPrivate               bool   `json:"isPrivate"`
	IsRestricted            bool   `json:"isRestricted"`
	IsBestie                bool   `json:"isBestie"`
	OutgoingRequest         bool   `json:"outgoingRequest"`
	IncomingRequest         bool   `json:"incomingRequest"`
	IsFeedFavorite          bool   `json:"isFeedFavorite,omitempty"`
}

// Hashtag represents a Threads hashtag with engagement counts.
type Hashtag struct {
	ID                  string `json:"id"`
	Name                string `json:"name"`
	MediaCount          int    `json:"mediaCount,omitempty"`
	ThreadsCount        int    `json:"threadsCount,omitempty"`
	UseInBio            bool   `json:"useInBio,omitempty"`
	FollowStatus        string `json:"followStatus,omitempty"`
	IsTrending          bool   `json:"isTrending,omitempty"`
	ProfilePicURL       string `json:"profilePicUrl,omitempty"`
	FormattedMediaCount string `json:"formattedMediaCount,omitempty"`
}

// HashtagFeed is the response from a hashtag posts endpoint.
type HashtagFeed struct {
	Hashtag    Hashtag  `json:"hashtag"`
	Threads    []Thread `json:"threads"`
	NextCursor string   `json:"nextCursor,omitempty"`
	HasNext    bool     `json:"hasNext"`
}

// Notification represents a single entry from the user's inbox.
type Notification struct {
	StoryType  int             `json:"storyType,omitempty"`
	Timestamp  time.Time       `json:"timestamp,omitempty"`
	Args       json.RawMessage `json:"args,omitempty"` // structure varies by story_type
}

// NotificationPage is one page of notifications.
type NotificationPage struct {
	Notifications []Notification `json:"notifications"`
	NextCursor    string         `json:"nextCursor,omitempty"`
	HasNext       bool           `json:"hasNext"`
}

// RateLimitState is the most recent rate-limit observation for the client.
//
// Threads shares Instagram's backend and does NOT publish standard
// X-RateLimit-* headers. Rate pressure is signalled via:
//   - "Please wait a few minutes before you try again." JSON body
//   - HTTP 429 responses
//   - Soft capacity hints in response headers (x-ig-*, x-fb-connection-quality)
//
// CooldownReadUntil / CooldownWriteUntil are set whenever a rate-limit signal
// is detected. All subsequent requests of the same kind block until the
// cooldown elapses. Use Client.RateLimit() to read, Client.WaitForCooldown()
// to block until clear.
type RateLimitState struct {
	// Cooldown tracks when the client may resume read / write requests.
	CooldownReadUntil  time.Time `json:"cooldownReadUntil,omitempty"`
	CooldownWriteUntil time.Time `json:"cooldownWriteUntil,omitempty"`

	// BlockedReason is a short human-readable description of why the last
	// cooldown was tripped (e.g. "HTTP 429", "wait-a-few-minutes", "302→login").
	BlockedReason string `json:"blockedReason,omitempty"`

	// LastBlockedAt is the timestamp of the most recent rate-limit event.
	LastBlockedAt time.Time `json:"lastBlockedAt,omitempty"`

	// WriteBlocked is true when the most recent rate-limit event was a write.
	WriteBlocked bool `json:"writeBlocked,omitempty"`

	// LastReadAt / LastWriteAt record the timestamp of the most recent
	// successful read / write request.
	LastReadAt  time.Time `json:"lastReadAt,omitempty"`
	LastWriteAt time.Time `json:"lastWriteAt,omitempty"`

	// Instagram / Meta soft-signal headers. These appear on most responses
	// from www.threads.com and i.instagram.com.
	//
	// CapacityLevel: 0 = degraded, 3 = healthy, -1 = not present in response.
	CapacityLevel       int    `json:"capacityLevel"`
	PeakTime            bool   `json:"peakTime,omitempty"`
	PeakV2              bool   `json:"peakV2,omitempty"`
	ConnectionQuality   string `json:"connectionQuality,omitempty"`
	OriginRegion        string `json:"originRegion,omitempty"`
	ServerRegion        string `json:"serverRegion,omitempty"`
	LastServerElapsedMs int    `json:"lastServerElapsedMs,omitempty"`
}

// PostOption configures CreatePost / Reply / Quote.
type PostOption func(*postOptions)

type postOptions struct {
	mediaIDs       []string
	imagePaths     []string // local image files to upload, then attach
	replyToID      string
	quotePostID    string
	replyControl   string // "everyone" | "accounts_you_follow" | "mentioned_only"
	publishMode    string // "text_post" (default) | other future modes
}

// WithReplyTo configures CreatePost to reply to a given post ID.
// Equivalent to calling Reply directly.
func WithReplyTo(postID string) PostOption {
	return func(o *postOptions) { o.replyToID = postID }
}

// WithQuote configures CreatePost to quote-repost a given post ID.
func WithQuote(postID string) PostOption {
	return func(o *postOptions) { o.quotePostID = postID }
}

// WithReplyControl restricts who can reply to the new post.
// Valid values: "everyone", "accounts_you_follow", "mentioned_only".
func WithReplyControl(scope string) PostOption {
	return func(o *postOptions) { o.replyControl = scope }
}

// WithImage attaches a local image file to the new post. Multiple images
// can be supplied to create a carousel. Each path is uploaded as JPEG.
func WithImage(path string) PostOption {
	return func(o *postOptions) { o.imagePaths = append(o.imagePaths, path) }
}

// WithMediaIDs attaches already-uploaded media to a new post. Use this if
// you've called UploadImage separately and want to control the upload.
func WithMediaIDs(ids ...string) PostOption {
	return func(o *postOptions) { o.mediaIDs = append(o.mediaIDs, ids...) }
}
