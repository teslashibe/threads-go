package threads

import (
	"encoding/json"
	"strings"
	"time"
)

// userPayload mirrors the Threads/Instagram user JSON object as returned by
// /accounts/current_user/, /users/{id}/info/, search results, etc. Only
// fields that map to the public User struct are included.
type userPayload struct {
	PK                          json.Number `json:"pk"`
	PKID                        string      `json:"pk_id"`
	StrongID                    string      `json:"strong_id__"`
	Username                    string      `json:"username"`
	FullName                    string      `json:"full_name"`
	IsPrivate                   bool        `json:"is_private"`
	IsVerified                  bool        `json:"is_verified"`
	ProfilePicURL               string      `json:"profile_pic_url"`
	HDProfilePicURLInfo         *imagePayload `json:"hd_profile_pic_url_info"`
	HDProfilePicVersions        []imagePayload `json:"hd_profile_pic_versions"`
	FollowerCount               int         `json:"follower_count"`
	FollowingCount              int         `json:"following_count"`
	MediaCount                  int         `json:"media_count"`
	FBIDV2                      json.Number `json:"fbid_v2"`
	HasNMEBadge                 bool        `json:"has_nme_badge"`
	ShowFBLinkOnProfile         bool        `json:"show_fb_link_on_profile"`
	ShowFBPageLinkOnProfile     bool        `json:"show_fb_page_link_on_profile"`
	ThirdPartyDownloadsEnabled  int         `json:"third_party_downloads_enabled"`
	TextAppIsLowLike            bool        `json:"text_app_is_low_like"`
	FeedPostReshareDisabled     bool        `json:"feed_post_reshare_disabled"`
	EligibleForActivBadge       bool        `json:"eligible_for_text_app_activation_badge"`
	HideActivBadgeOnTextApp     bool        `json:"hide_text_app_activation_badge_on_text_app"`
	TextAppBiography            *bioPayload `json:"text_app_biography"`
	BiographyWithEntities       *bioPayload `json:"biography_with_entities"`
	Biography                   string      `json:"biography"`
	ExternalURL                 string      `json:"external_url"`
	BioLinks                    []bioLink   `json:"bio_links"`
	FriendshipStatus            *friendshipPayload `json:"friendship_status"`
}

type bioPayload struct {
	TextFragments struct {
		Fragments []struct {
			Plaintext string `json:"plaintext"`
		} `json:"fragments"`
	} `json:"text_fragments"`
	RawText string `json:"raw_text"`
	Text    string `json:"text"`
}

type bioLink struct {
	URL    string `json:"url"`
	Title  string `json:"title"`
}

type imagePayload struct {
	URL    string `json:"url"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
}

type videoPayload struct {
	URL    string `json:"url"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
	Type   int    `json:"type"`
}

type friendshipPayload struct {
	UserID                   json.Number `json:"user_id"`
	Following                bool        `json:"following"`
	FollowedBy               bool        `json:"followed_by"`
	Blocking                 bool        `json:"blocking"`
	Muting                   bool        `json:"muting"`
	IsPrivate                bool        `json:"is_private"`
	IsRestricted             bool        `json:"is_restricted"`
	IsBestie                 bool        `json:"is_bestie"`
	OutgoingRequest          bool        `json:"outgoing_request"`
	IncomingRequest          bool        `json:"incoming_request"`
	IsFeedFavorite           bool        `json:"is_feed_favorite"`
}

// postPayload mirrors a single thread item / media JSON object.
type postPayload struct {
	PK                       json.Number  `json:"pk"`
	ID                       string       `json:"id"`
	StrongID                 string       `json:"strong_id__"`
	FBID                     string       `json:"fbid"`
	Code                     string       `json:"code"`
	TakenAt                  int64        `json:"taken_at"`
	MediaType                int          `json:"media_type"`
	ProductType              string       `json:"product_type"`
	LikeCount                int          `json:"like_count"`
	HasLiked                 bool         `json:"has_liked"`
	CanViewerReshare         bool         `json:"can_viewer_reshare"`
	CanReply                 bool         `json:"can_reply"`
	IsReplyToAuthor          bool         `json:"is_reply_to_author_of_root_thread"`
	Caption                  *captionPayload `json:"caption"`
	User                     userPayload  `json:"user"`
	ImageVersions2           *struct {
		Candidates []imagePayload `json:"candidates"`
	} `json:"image_versions2"`
	VideoVersions            []videoPayload `json:"video_versions"`
	CarouselMedia            []postPayload  `json:"carousel_media"`
	CaptionIsEdited          bool         `json:"caption_is_edited"`
	OrganicTrackingToken     string       `json:"organic_tracking_token"`
	OriginalWidth            int          `json:"original_width"`
	OriginalHeight           int          `json:"original_height"`
	TextPostAppInfo          *textPostInfo `json:"text_post_app_info"`
}

type captionPayload struct {
	Text      string `json:"text"`
	CreatedAt int64  `json:"created_at"`
}

type textPostInfo struct {
	DirectReplyCount int    `json:"direct_reply_count"`
	RepostCount      int    `json:"repost_count"`
	QuoteCount       int    `json:"quote_count"`
	LinkPreviewAttachment *struct {
		URL         string `json:"url"`
		DisplayURL  string `json:"display_url"`
		Title       string `json:"title"`
		ImageURL    string `json:"image_url"`
	} `json:"link_preview_attachment"`
	ShareInfo *struct {
		QuotedPost   *postPayload `json:"quoted_post"`
		RepostedPost *postPayload `json:"reposted_post"`
	} `json:"share_info"`
	IsPostUnavailable bool `json:"is_post_unavailable"`
	ReplyToAuthor     *userPayload `json:"reply_to_author"`
}

// threadPayload mirrors the thread wrapper used in feed responses.
//
// thread_type/thread_item_type are kept as RawMessage because Meta
// returns them as either int (legacy) or string (newer feeds).
type threadPayload struct {
	ID          string          `json:"id"`
	ThreadType  json.RawMessage `json:"thread_type"`
	ThreadItems []threadItem    `json:"thread_items"`
}

type threadItem struct {
	Post           postPayload     `json:"post"`
	ThreadItemType json.RawMessage `json:"thread_item_type"`
}

// hashtagPayload mirrors a hashtag JSON object.
type hashtagPayload struct {
	ID                  json.Number `json:"id"`
	Name                string      `json:"name"`
	MediaCount          int         `json:"media_count"`
	ThreadsCount        int         `json:"threads_count"`
	UseInBio            bool        `json:"use_in_bio"`
	FollowStatus        string      `json:"follow_status"`
	IsTrending          bool        `json:"is_trending"`
	ProfilePicURL       string      `json:"profile_pic_url"`
	FormattedMediaCount string      `json:"formatted_media_count"`
}

// ----------------------------------------------------------------------------
// Converters: payload -> public type
// ----------------------------------------------------------------------------

func toUser(p userPayload) User {
	id := stringID(p.PK)
	if id == "" {
		id = p.PKID
	}
	if id == "" {
		id = p.StrongID
	}
	u := User{
		ID:                       id,
		FBID:                     stringID(p.FBIDV2),
		Username:                 p.Username,
		FullName:                 p.FullName,
		IsPrivate:                p.IsPrivate,
		IsVerified:               p.IsVerified,
		ProfilePicURL:            p.ProfilePicURL,
		FollowerCount:            p.FollowerCount,
		FollowingCount:           p.FollowingCount,
		MediaCount:               p.MediaCount,
		HasNMEBadge:              p.HasNMEBadge,
		ShowFBLinkOnProfile:      p.ShowFBLinkOnProfile,
		ThirdPartyDownloads:      p.ThirdPartyDownloadsEnabled,
		TextAppLowLike:           p.TextAppIsLowLike,
		FeedPostReshareDisabled:  p.FeedPostReshareDisabled,
		EligibleForActivBadge:    p.EligibleForActivBadge,
		HideActivBadgeOnTextApp:  p.HideActivBadgeOnTextApp,
		ExternalURL:              p.ExternalURL,
	}
	if p.HDProfilePicURLInfo != nil {
		u.HDProfilePicURL = p.HDProfilePicURLInfo.URL
	}
	for _, v := range p.HDProfilePicVersions {
		u.HDProfilePicVersions = append(u.HDProfilePicVersions, Image{URL: v.URL, Width: v.Width, Height: v.Height})
	}
	if bio := extractBio(p); bio != "" {
		u.Biography = bio
	}
	for _, link := range p.BioLinks {
		if link.URL != "" {
			u.BiographyLinks = append(u.BiographyLinks, link.URL)
		}
	}
	return u
}

func extractBio(p userPayload) string {
	if p.TextAppBiography != nil {
		var parts []string
		for _, frag := range p.TextAppBiography.TextFragments.Fragments {
			if frag.Plaintext != "" {
				parts = append(parts, frag.Plaintext)
			}
		}
		if len(parts) > 0 {
			return strings.Join(parts, "")
		}
		if p.TextAppBiography.RawText != "" {
			return p.TextAppBiography.RawText
		}
		if p.TextAppBiography.Text != "" {
			return p.TextAppBiography.Text
		}
	}
	if p.BiographyWithEntities != nil {
		if p.BiographyWithEntities.RawText != "" {
			return p.BiographyWithEntities.RawText
		}
	}
	return p.Biography
}

func toPost(p postPayload) Post {
	id := stringID(p.PK)
	if id == "" {
		id = p.ID
	}
	if id == "" {
		id = p.StrongID
	}
	post := Post{
		ID:                  id,
		FBID:                p.FBID,
		Code:                p.Code,
		MediaType:           p.MediaType,
		ProductType:         p.ProductType,
		LikeCount:           p.LikeCount,
		HasLiked:            p.HasLiked,
		CanReshare:          p.CanViewerReshare,
		CanReply:            p.CanReply,
		IsReplyToAuthor:     p.IsReplyToAuthor,
		User:                toUser(p.User),
		CaptionEdited:       p.CaptionIsEdited,
		OrganicTracking:     p.OrganicTrackingToken,
		OriginalWidth:       p.OriginalWidth,
		OriginalHeight:      p.OriginalHeight,
	}
	if p.TakenAt > 0 {
		post.TakenAt = time.Unix(p.TakenAt, 0).UTC()
	}
	if p.Caption != nil {
		post.Text = p.Caption.Text
	}
	if p.ImageVersions2 != nil {
		for _, c := range p.ImageVersions2.Candidates {
			post.ImageVersions = append(post.ImageVersions, Image{URL: c.URL, Width: c.Width, Height: c.Height})
		}
	}
	for _, v := range p.VideoVersions {
		post.VideoVersions = append(post.VideoVersions, Video{URL: v.URL, Width: v.Width, Height: v.Height, Type: v.Type})
	}
	for _, cm := range p.CarouselMedia {
		post.CarouselMedia = append(post.CarouselMedia, toPost(cm))
	}
	if p.TextPostAppInfo != nil {
		post.ReplyCount = p.TextPostAppInfo.DirectReplyCount
		post.RepostCount = p.TextPostAppInfo.RepostCount
		post.QuoteCount = p.TextPostAppInfo.QuoteCount
		if lp := p.TextPostAppInfo.LinkPreviewAttachment; lp != nil && lp.URL != "" {
			post.LinkPreview = &LinkPreview{
				URL:        lp.URL,
				DisplayURL: lp.DisplayURL,
				Title:      lp.Title,
				ImageURL:   lp.ImageURL,
			}
		}
		if si := p.TextPostAppInfo.ShareInfo; si != nil {
			if si.QuotedPost != nil {
				qp := toPost(*si.QuotedPost)
				post.QuotedPost = &qp
			}
			if si.RepostedPost != nil {
				rp := toPost(*si.RepostedPost)
				post.RepostedPost = &rp
			}
		}
	}
	if post.User.Username != "" && post.Code != "" {
		post.Permalink = "https://www.threads.com/@" + post.User.Username + "/post/" + post.Code
	}
	return post
}

func toThread(p threadPayload) Thread {
	t := Thread{ID: p.ID}
	for _, item := range p.ThreadItems {
		t.ThreadItems = append(t.ThreadItems, toPost(item.Post))
	}
	return t
}

func toThreads(ps []threadPayload) []Thread {
	out := make([]Thread, 0, len(ps))
	for _, p := range ps {
		out = append(out, toThread(p))
	}
	return out
}

func toFriendship(p *friendshipPayload) FriendshipStatus {
	if p == nil {
		return FriendshipStatus{}
	}
	return FriendshipStatus{
		UserID:          stringID(p.UserID),
		Following:       p.Following,
		FollowedBy:      p.FollowedBy,
		Blocking:        p.Blocking,
		Muting:          p.Muting,
		IsPrivate:       p.IsPrivate,
		IsRestricted:    p.IsRestricted,
		IsBestie:        p.IsBestie,
		OutgoingRequest: p.OutgoingRequest,
		IncomingRequest: p.IncomingRequest,
		IsFeedFavorite:  p.IsFeedFavorite,
	}
}

func toHashtag(p hashtagPayload) Hashtag {
	return Hashtag{
		ID:                  stringID(p.ID),
		Name:                p.Name,
		MediaCount:          p.MediaCount,
		ThreadsCount:        p.ThreadsCount,
		UseInBio:            p.UseInBio,
		FollowStatus:        p.FollowStatus,
		IsTrending:          p.IsTrending,
		ProfilePicURL:       p.ProfilePicURL,
		FormattedMediaCount: p.FormattedMediaCount,
	}
}

func stringID(n json.Number) string {
	s := n.String()
	if s == "" || s == "0" {
		return ""
	}
	return s
}
