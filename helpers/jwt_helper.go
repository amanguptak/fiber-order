package helpers

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const SecretKey = "secret"

func GenerateToken(issuer string, expirationTime time.Time) (string, error) {
	claims := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"iss": issuer,
		"exp": expirationTime.Unix(),
	})

	return claims.SignedString([]byte(SecretKey))
}

func ParseToken(tokenString string) (*jwt.Token, error) {
	return jwt.ParseWithClaims(tokenString, &jwt.MapClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(SecretKey), nil
	})
}


// This function is the "Security Guard" of your app. It checks if a token is Real or Fake.

// Here is the breakdown of what is happening inside 
// ParseToken
// :

// The Code
// go
// func ParseToken(tokenString string) (*jwt.Token, error) {
//     return jwt.ParseWithClaims(tokenString, &jwt.MapClaims{}, func(token *jwt.Token) (interface{}, error) {
//         return []byte(SecretKey), nil
//     })
// }
// 1. What is jwt.ParseWithClaims doing?
// It takes the messy string eyJhbG... (the token) and tries to turn it back into readable data (JSON). But it doesn't just trust the data; it verifies it.

// It performs 3 steps automatically:

// Decodes: Converts the Base64 string back to JSON.
// Checks Expiration: If the exp time in the token is in the past, it immediately fails and returns an error. You don't even need to write code for this!
// Verifies Signature: This is the most important part. It checks if the token was signed by YOU (using your SecretKey).
// 2. What is that weird function inside?
// go
// func(token *jwt.Token) (interface{}, error) {
//     return []byte(SecretKey), nil
// }
// This is a Callback Function.

// The Problem: The JWT library doesn't know your secret password (SecretKey).
// The Solution: It asks you: "Hey, I found this token. What secret key should I use to check if it's valid?"
// Your Answer: You return []byte(SecretKey).
// Why is this a function? In complex apps, you might have multiple keys (e.g., "Key A" for users, "Key B" for admins). This function lets you look at the token first and decide which key to give back. For us, we just always return the same SecretKey.

// 3. What is it returning?
// It returns two things:

// *jwt.Token: This is an object containing all the data.
// token.Valid: A boolean (true or false). This is what you check.
// token.Claims: The actual data (User ID, Email) inside the token.
// error: If the token is expired, fake, or malformed, this will be not nil.
// Summary
// You give it a String (token). It uses your SecretKey to check the math. It returns a Token Object (valid/invalid) and the Data inside it.





/////////////***************************--------------------




//CallBack Function in golang


// Yes! In Go, functions are "First-Class Citizens".

// This means:

// You can save a function in a variable.
// You can pass a function as an argument to another function (this is a Callback).
// You can return a function from another function.
// Example: A Simple Callback in Go
// Here is a simple example to help you visualize it. Imagine a function that filters numbers.

// go
// package main

// import "fmt"

// // 1. We define a function that TAKES another function as an argument
// // 'check' is the callback function. It takes an int and returns a bool.
// func FilterNumbers(numbers []int, check func(int) bool) []int {
//     var result []int
//     for _, n := range numbers {
//         // 2. We CALL the callback function here
//         if check(n) {
//             result = append(result, n)
//         }
//     }
//     return result
// }

// func main() {
//     nums := []int{1, 2, 3, 4, 5, 6}

//     // 3. We pass an ANONYMOUS function (lambda) as the callback
//     evens := FilterNumbers(nums, func(n int) bool {
//         return n%2 == 0 // Return true if even
//     })

//     fmt.Println(evens) // Output: [2 4 6]
// }
// This is exactly what jwt.ParseWithClaims is doing. It's asking you to provide a small function logic (func(token)) that it can use internally during its process.





// The Code
// go
// jwt.ParseWithClaims(tokenString, &jwt.MapClaims{}, func(token *jwt.Token) (interface{}, error) { ... })
// 1. &jwt.MapClaims{} (The "Empty Bucket")
// What is it? jwt.MapClaims is just a map[string]interface{}. It's a place to store key-value pairs (like "email": "
// aman@test.com
// ").
// Why & (Pointer)? We are passing the Address of an empty map.
// Why? The ParseWithClaims function needs a place to put the data it finds inside the token. By giving it a pointer (address), we are saying: "Here is an empty bucket. Please fill it up with the data you find in the token."
// 2. func(token *jwt.Token) (The Callback Input)
// What is it? This is an anonymous function (a function without a name) that we are defining right here.
// token *jwt.Token: The library calls this function and passes us the token object it just parsed. This allows us to inspect the token (e.g., check which signing method was used) before we decide to give the key.
// 3. 
// (interface{}, error)
//  (The Callback Return Types)
// This is the part you asked about. Why interface{}?

// The Problem: There are many ways to sign a JWT:
// HS256: Uses a simple password (bytes) like "secret".
// RS256: Uses a Public/Private Key pair (RSA objects).
// ES256: Uses Elliptic Curve keys (ECDSA objects).
// The Go Constraint: Go is a "Strongly Typed" language. A function usually has to return a specific type, like string or int. You can't usually say "Return a String OR an RSA Key".
// The Solution (interface{}):
// interface{} is a special type in Go that means "Anything".
// It is a container that can hold any value: a string, a number, a struct, or a pointer.
// The library authors used interface{} here because they don't know what kind of encryption you are using. They are saying: "Return whatever key object is correct for your algorithm, and I will figure out how to use it."
// 4. return []byte(SecretKey), nil
// SecretKey: This is your string "secret".
// []byte(...): We convert the string into a "Slice of Bytes".
// Why? The HMAC algorithm (HS256) works on raw bytes, not strings. So we must convert it.
// Since []byte satisfies interface{}, this is a valid return value.
// nil: This is the error return. We are saying "No error, here is the key."
// Summary in Plain English
// You: "Hey Library, please parse this tokenString."
//  You: "Here is an empty bucket (&jwt.MapClaims{}). Put the data in there."
//  You: "And here is a helper function. If you need the secret key to check 
// the signature, call this function." Library: (Calls your function) 
// "I need the key." Your Function: "Here it is as bytes ([]byte)." (This is returned as an interface{}). Library: "Thanks. I verified the signature. It matches. Here is the data."