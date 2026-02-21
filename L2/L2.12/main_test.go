package main

import (
	"regexp"
	"testing"
)

func TestMatchFixed(t *testing.T) {
	opts := &GrepOptions{Fixed: true}

	opts.Pattern = "hello"
	result := match("hello world", opts, nil)
	if !result {
		t.Errorf("Expected match for 'hello' in 'hello world'")
	}

	opts.Pattern = "goodbye"
	result = match("hello world", opts, nil)
	if result {
		t.Errorf("Expected no match for 'goodbye' in 'hello world'")
	}
}

func TestMatchFixedIgnoreCase(t *testing.T) {
	opts := &GrepOptions{Fixed: true, IgnoreCase: true}

	opts.Pattern = "hello"
	result := match("HELLO world", opts, nil)
	if !result {
		t.Errorf("Expected case-insensitive match for 'hello' in 'HELLO world'")
	}
}

func TestMatchInvert(t *testing.T) {
	opts := &GrepOptions{Fixed: true, Invert: true}

	opts.Pattern = "hello"
	result := match("hello world", opts, nil)
	if result {
		t.Errorf("Expected inverted match to return false for matching line")
	}

	result = match("goodbye world", opts, nil)
	if !result {
		t.Errorf("Expected inverted match to return true for non-matching line")
	}
}

func TestMatchRegexp(t *testing.T) {
	opts := &GrepOptions{Fixed: false}

	regex, _ := regexp.Compile("[0-9]+")
	opts.Pattern = "[0-9]+"

	result := match("abc123", opts, regex)
	if !result {
		t.Errorf("Expected regex match for '[0-9]+' in 'abc123'")
	}

	result = match("abcdef", opts, regex)
	if result {
		t.Errorf("Expected no regex match for '[0-9]+' in 'abcdef'")
	}
}

func TestMatchRegexpIgnoreCase(t *testing.T) {
	opts := &GrepOptions{Fixed: false, IgnoreCase: true}

	regex, _ := regexp.Compile("(?i)hello")
	opts.Pattern = "hello"

	result := match("HELLO world", opts, regex)
	if !result {
		t.Errorf("Expected case-insensitive regex match")
	}
}
