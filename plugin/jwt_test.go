package plugin

import (
	"fmt"
	"strings"
	"testing"

	"github.com/extrame/jose/jws"
)

func TestJwt(t *testing.T) {
	var claims = make(jws.Claims)
	claims.Set("userId", "1")
	method := jws.GetSigningMethod("HS512")
	j := jws.NewJWT(claims, method)
	j.Claims().SetIssuer("test")

	b, err := j.Serialize([]byte("test"))
	if err == nil {
		fmt.Printf("Bearer %s\n", string(b))
		auth := strings.TrimPrefix(string(b), "Bearer ")
		token, err := jws.ParseJWT([]byte(auth))
		if err == nil {
			err = token.Validate([]byte("test"))
			if err == nil {
				fmt.Println("Result is:", token.Claims().Get("userId").(string))
			} else {
				t.Fatal(err)
			}
		} else {
			t.Fatal(err)
		}
	} else {
		t.Fatal(err)
	}
}
