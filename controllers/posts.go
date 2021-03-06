package controllers

import (
	"net/http"
	"strconv"

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
	Text string `json:"text" binding:"required"`
}

type AddPostReturn struct {
	ID    uint   `json:"id"`
	Token string `json:"token"`
}

func GetPosts(c *gin.Context) {
	var posts []ReturnPost
	tokens := c.GetStringSlice("tokens")
	limit, _ := strconv.Atoi(c.Query("limit"))
	if limit < 20 {
		limit = 20
	}
	offset := limit - 20
	rows, err := models.DB.Query("SELECT * FROM posts ORDER BY id DESC LIMIT $1 OFFSET $2", limit, offset)
	if err != nil {
		panic(err)
	}
	for rows.Next() {
		newPost := ReturnPost{
			CanUserModify: false,
		}
		var secretId string
		rows.Scan(&newPost.ID, &secretId, &newPost.Text)
		for _, token := range tokens {
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
	tokens := c.GetStringSlice("tokens")
	var editedPost PostFromUser
	if err := c.BindJSON(&editedPost); err != nil {
		c.Status(http.StatusBadRequest)
		return
	}
	id := c.Param("id")
	if len(tokens) == 0 {
		c.Status(http.StatusUnauthorized)
		return
	}

	response, err := models.DB.Exec(`UPDATE posts SET text=$1 WHERE id=$2 AND secret_id = ANY($3)`, editedPost.Text, id, pq.Array(tokens))
	rowsAffected, _ := response.RowsAffected()
	if err != nil || rowsAffected == 0 {
		c.Status(http.StatusNotFound)
		return
	}
	c.Status(http.StatusOK)
}

func RemovePost(c *gin.Context) {
	tokens := c.GetStringSlice("tokens")
	id := c.Param("id")

	if len(tokens) == 0 || id == "" {
		c.Status(http.StatusBadRequest)
		return
	}

	response, err := models.DB.Exec("DELETE FROM posts WHERE id=$1 AND secret_id = ANY($2)", id, pq.Array(tokens))
	rowsAffected, _ := response.RowsAffected()
	if err != nil || rowsAffected == 0 {
		c.Status(http.StatusNotFound)
		return
	}
	c.Status(http.StatusOK)
}
