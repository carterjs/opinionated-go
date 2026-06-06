package javagetters

type User struct {
	name string
	age  int
}

// Bad: Java-style getter
func (u *User) GetName() string { // want `method "GetName" should be named "name"`
	return u.name
}

// Bad: Java-style getter
func (u *User) GetAge() int { // want `method "GetAge" should be named "age"`
	return u.age
}

// Good: Go idiom
func (u *User) Name() string {
	return u.name
}

// Good: Go idiom
func (u *User) Age() int {
	return u.age
}

// Good: Setter is acceptable
func (u *User) SetName(name string) {
	u.name = name
}

// Good: Unexported methods don't need to follow this rule
func (u *User) getName() string {
	return u.name
}
