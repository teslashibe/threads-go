package mcp

import (
	"context"

	"github.com/teslashibe/mcptool"
	threads "github.com/teslashibe/threads-go"
)

// CreatePostInput is the typed input for threads_create_post.
//
// Pass at most one of ReplyToID or QuotePostID. ImagePath, when provided,
// is uploaded as JPEG via UploadImage and attached to the new post.
type CreatePostInput struct {
	Text        string `json:"text" jsonschema:"description=plain-text post body,required"`
	ReplyToID   string `json:"reply_to_id,omitempty" jsonschema:"description=if set, publish as a reply to this post ID (mutually exclusive with quote_post_id)"`
	QuotePostID string `json:"quote_post_id,omitempty" jsonschema:"description=if set, publish as a quote-repost of this post ID (mutually exclusive with reply_to_id)"`
	ImagePath   string `json:"image_path,omitempty" jsonschema:"description=optional local filesystem path to a JPEG to attach"`
}

func buildPostOpts(in CreatePostInput) []threads.PostOption {
	var opts []threads.PostOption
	if in.ReplyToID != "" {
		opts = append(opts, threads.WithReplyTo(in.ReplyToID))
	}
	if in.QuotePostID != "" {
		opts = append(opts, threads.WithQuote(in.QuotePostID))
	}
	if in.ImagePath != "" {
		opts = append(opts, threads.WithImage(in.ImagePath))
	}
	return opts
}

func createPost(ctx context.Context, c *threads.Client, in CreatePostInput) (any, error) {
	return c.CreatePost(ctx, in.Text, buildPostOpts(in)...)
}

// ReplyInput is the typed input for threads_reply.
type ReplyInput struct {
	PostID    string `json:"post_id" jsonschema:"description=numeric post ID to reply to,required"`
	Text      string `json:"text" jsonschema:"description=plain-text reply body,required"`
	ImagePath string `json:"image_path,omitempty" jsonschema:"description=optional local filesystem path to a JPEG to attach"`
}

func reply(ctx context.Context, c *threads.Client, in ReplyInput) (any, error) {
	var opts []threads.PostOption
	if in.ImagePath != "" {
		opts = append(opts, threads.WithImage(in.ImagePath))
	}
	return c.Reply(ctx, in.PostID, in.Text, opts...)
}

// QuoteInput is the typed input for threads_quote.
type QuoteInput struct {
	PostID    string `json:"post_id" jsonschema:"description=numeric post ID to quote-repost,required"`
	Text      string `json:"text" jsonschema:"description=plain-text body for the quote post,required"`
	ImagePath string `json:"image_path,omitempty" jsonschema:"description=optional local filesystem path to a JPEG to attach"`
}

func quote(ctx context.Context, c *threads.Client, in QuoteInput) (any, error) {
	var opts []threads.PostOption
	if in.ImagePath != "" {
		opts = append(opts, threads.WithImage(in.ImagePath))
	}
	return c.Quote(ctx, in.PostID, in.Text, opts...)
}

// UploadImageInput is the typed input for threads_upload_image.
type UploadImageInput struct {
	Path string `json:"path" jsonschema:"description=local filesystem path to a JPEG to upload,required"`
}

func uploadImage(ctx context.Context, c *threads.Client, in UploadImageInput) (any, error) {
	id, err := c.UploadImage(ctx, in.Path)
	if err != nil {
		return nil, err
	}
	return map[string]any{"upload_id": id}, nil
}

var postTools = []mcptool.Tool{
	mcptool.Define[*threads.Client, CreatePostInput](
		"threads_create_post",
		"Publish a new Threads post, optionally as a reply, quote, or with an image",
		"CreatePost",
		createPost,
	),
	mcptool.Define[*threads.Client, ReplyInput](
		"threads_reply",
		"Reply to a Threads post",
		"Reply",
		reply,
	),
	mcptool.Define[*threads.Client, QuoteInput](
		"threads_quote",
		"Quote-repost a Threads post with commentary",
		"Quote",
		quote,
	),
	mcptool.Define[*threads.Client, UploadImageInput](
		"threads_upload_image",
		"Upload a local JPEG image and return the upload_id for later attachment",
		"UploadImage",
		uploadImage,
	),
}
