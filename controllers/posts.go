package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/tell-something-nice-backend/models"

	"github.com/lib/pq"
)

type ReturnPost struct {
	ID            uint   `json:"id"`
	Text          string `json:"text"`
	CanUserModify bool   `json:"canUserModify"`
}

type PostFromUser struct {
	Text string `json:"text"`
}

type AddPostReturn struct {
	ID    uint   `json:"id"`
	Token string `json:"token"`
}

func GetPosts(c *gin.Context) {
	var posts []ReturnPost
	tokens, _ := c.Get("tokens")
	tokensArr := tokens.([]string)

	rows, err := models.DB.Query("SELECT * FROM posts")
	if err != nil {
		panic(err)
	}
	for rows.Next() {
		newPost := ReturnPost{
			CanUserModify: false,
		}
		var secretId string
		rows.Scan(&newPost.ID, &secretId, &newPost.Text)
		for _, token := range tokensArr {
			if token == secretId {
				newPost.CanUserModify = true
			}
		}
		posts = append(posts, newPost)
	}
	c.JSON(http.StatusOK, posts)
}

func AddPost(c *gin.Context) {
	post := PostFromUser{}
	if err := c.BindJSON(&post); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Your data is not correct"})
	}
	data := AddPostReturn{
		Token: uuid.New().String(),
	}
	err := models.DB.QueryRow("INSERT INTO posts(text, secret_id) VALUES ($1, $2) RETURNING id", post.Text, data.Token).Scan(&data.ID)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}
	c.JSON(http.StatusOK, data)
}

func EditPost(c *gin.Context) {
	tokens, _ := c.Get("tokens")
	var editedPost PostFromUser
	if err := c.BindJSON(&editedPost); err != nil {
		c.Status(http.StatusBadRequest)
		return
	}
	tokensArr := tokens.([]string)
	id := c.Param("id")
	if len(tokensArr) == 0 || id == "" {
		c.Status(http.StatusBadRequest)
		return
	}
	response, err := models.DB.Exec(`UPDATE posts SET text=$1 WHERE id=$2 AND secret_id = ANY($3)`, editedPost.Text, id, pq.Array(tokensArr))
	rowsAffected, _ := response.RowsAffected()
	if err != nil || rowsAffected == 0 {
		c.Status(http.StatusNotFound)
		return
	}
	c.Status(http.StatusOK)
}

func RemovePost(c *gin.Context) {
	tokens, _ := c.Get("tokens")
	tokensArr := tokens.([]string)
	id := c.Param("id")

	if len(tokensArr) == 0 || id == "" {
		c.Status(http.StatusBadRequest)
		return
	}

	response, err := models.DB.Exec("DELETE FROM posts WHERE id=$1 AND secret_id = ANY($2)", id, pq.Array(tokensArr))
	rowsAffected, _ := response.RowsAffected()
	if err != nil || rowsAffected == 0 {
		c.Status(http.StatusNotFound)
		return
	}
	c.Status(http.StatusOK)
}
