package singleletterexported

type X struct{} // want "exported name .* is too short"

type Message struct{}

func Y() {} // want "exported name .* is too short"

func Process() {}

const Z = 1 // want "exported name .* is too short"

const MaxSize = 100
