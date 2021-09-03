package render

import "forum/pkg/model"

func CreateAllForum(in []*model.Forum) []*model.ForumResponse {
	if len(in) < 1 {
		return []*model.ForumResponse{}
	}

	out := make([]*model.ForumResponse, len(in))
	for i, forum := range in {
		out[i] = &model.ForumResponse{
			ForumID:   forum.ForumId,
			Name:      forum.Name,
			Intro:     forum.Intro,
			Sort:      forum.Sort,
			Parent:    forum.Parent,
			CreatedAt: forum.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt: forum.UpdatedAt.Format("2006-01-02 15:04:05"),
		}
	}
	return out
}
