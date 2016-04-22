package twitter

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"testing"
)

func TestPost_TwitterPostConv(t *testing.T) {
	b, _ := ioutil.ReadFile("./_testdata/twitter-post.json")
	buffer := bytes.NewReader(b)
	decoder := json.NewDecoder(buffer)

	twitterPost := TwitterPost{}
	if err := decoder.Decode(&twitterPost); err != nil {
		t.Error(err)
	}

	post := FromTwitterPost(twitterPost)

	if post.Id != 712860586267316224 {
		t.Errorf("Ids do not match: %d != %d", 712860586267316224, post.Id)
	}

	if post.UserId != 9464552 {
		t.Errorf("UserIds do not match: %d != %d", 9464552, post.UserId)
	}

	if len(post.Urls) != 1 {
		t.Errorf("Unexpected num of urls: %d != %d", 1, len(post.Urls))
	}

	if post.Retweet != nil {
		t.Errorf("Unexpected retweet")
	}

	if retweet, err := post.Stat.Get("retweet"); err != nil {
		t.Error(err)
	} else if retweet != 3 {
		t.Errorf("Unexpected retweet count: %d != %f", 3, retweet)
	}

	if favorite, err := post.Stat.Get("favorite"); err != nil {
		t.Error(err)
	} else if favorite != 3 {
		t.Errorf("Unexpected favorite count: %d != %f", 3, favorite)
	}

}

func TestPost_Json(t *testing.T) {
	writePosts := []Post{
		Post{Id: 100},
		Post{Id: 200},
		Post{Id: 300},
	}

	writeBuf := bytes.NewBuffer([]byte{})
	encoder := json.NewEncoder(writeBuf)

	if err := encoder.Encode(writePosts); err != nil {
		t.Error(err)
	}

	jsonb := writeBuf.Bytes()

	readBuf := bytes.NewBuffer(jsonb)
	decoder := json.NewDecoder(readBuf)

	var readPosts []Post
	if err := decoder.Decode(&readPosts); err != nil {
		t.Error(err)
	}

	if numRead := len(readPosts); numRead != 3 {
		t.Error("Unexpected number of posts read: %d != %d", 3, len(readPosts))
	}

	if firstPost := readPosts[0]; firstPost.Id != 100 {
		t.Error("Unexpected post id: %d != %d", 100, firstPost.Id)
	}

	if secondPost := readPosts[1]; secondPost.Id != 200 {
		t.Error("Unexpected post id: %d != %d", 200, secondPost.Id)
	}

	if thirdPost := readPosts[2]; thirdPost.Id != 300 {
		t.Error("Unexpected post id: %d != %d", 300, thirdPost.Id)
	}
}
