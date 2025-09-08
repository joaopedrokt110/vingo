package vingo

import (
	"regexp"
	"strings"
)

type TokenType int

const (
	TText TokenType = iota
	TVar
	TIf
	TElseIf
	TElse
	TEndIf
	TFor
	TEndFor
	TSwitch
	TCase
	TDefault
	TEndSwitch
)

type Token struct {
	Type    TokenType
	Value   string // for Var: expression or name; for If/For/Switch/Case: expression / raw
	Default string // for Var default literal (if provided)
	Raw     string // raw tag text
}

var (
	varPattern       = regexp.MustCompile(`^\s*(\w+(?:\.\w+)*)(?:\s*\|\s*"(.*?)")?\s*$`)
	ifPattern        = regexp.MustCompile(`^if\s+(.+)$`)
	elseifPattern    = regexp.MustCompile(`^elseif\s+(.+)$`)
	elsePattern      = regexp.MustCompile(`^else$`)
	endifPattern     = regexp.MustCompile(`^/if$`)
	forPattern       = regexp.MustCompile(`^for\s+(.+)\s+in\s+(.+)$`)
	endforPattern    = regexp.MustCompile(`^/for$`)
	switchPattern    = regexp.MustCompile(`^switch\s+(.+)$`)
	casePattern      = regexp.MustCompile(`^case\s+(.+)$`)
	defaultPattern   = regexp.MustCompile(`^default$`)
	endswitchPattern = regexp.MustCompile(`^/switch$`)
)

func tokenize(input string) []*Token {
	var tokens []*Token

	// Her parçayı ayır ve boş olanları atla
	parts := strings.Split(input, "<{")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		// Eğer kapanış "}>” yoksa text olarak ekle, baştaki <{ işaretini kaldırıyoruz
		sub := strings.SplitN(part, "}>", 2)
		if len(sub) == 2 {
			tag := strings.TrimSpace(sub[0])
			rest := sub[1]

			switch {
			case ifPattern.MatchString(tag):
				m := ifPattern.FindStringSubmatch(tag)
				tokens = append(tokens, &Token{Type: TIf, Value: m[1], Raw: tag})
			case elseifPattern.MatchString(tag):
				m := elseifPattern.FindStringSubmatch(tag)
				tokens = append(tokens, &Token{Type: TElseIf, Value: m[1], Raw: tag})
			case elsePattern.MatchString(tag):
				tokens = append(tokens, &Token{Type: TElse, Raw: tag})
			case endifPattern.MatchString(tag):
				tokens = append(tokens, &Token{Type: TEndIf, Raw: tag})
			case forPattern.MatchString(tag):
				m := forPattern.FindStringSubmatch(tag)
				tokens = append(tokens, &Token{Type: TFor, Value: strings.TrimSpace(m[1]) + ":" + strings.TrimSpace(m[2]), Raw: tag})
			case endforPattern.MatchString(tag):
				tokens = append(tokens, &Token{Type: TEndFor, Raw: tag})
			case switchPattern.MatchString(tag):
				m := switchPattern.FindStringSubmatch(tag)
				tokens = append(tokens, &Token{Type: TSwitch, Value: m[1], Raw: tag})
			case casePattern.MatchString(tag):
				m := casePattern.FindStringSubmatch(tag)
				tokens = append(tokens, &Token{Type: TCase, Value: m[1], Raw: tag})
			case defaultPattern.MatchString(tag):
				tokens = append(tokens, &Token{Type: TDefault, Raw: tag})
			case endswitchPattern.MatchString(tag):
				tokens = append(tokens, &Token{Type: TEndSwitch, Raw: tag})
			case varPattern.MatchString(tag):
				m := varPattern.FindStringSubmatch(tag)
				tokens = append(tokens, &Token{Type: TVar, Value: m[1], Default: m[2], Raw: tag})
			default:
				// bilinmeyen tag text olarak bırak
				tokens = append(tokens, &Token{Type: TText, Value: rest})
			}

			// Tag sonrası kalan text varsa ekle
			if rest != "" {
				tokens = append(tokens, &Token{Type: TText, Value: rest})
			}
		} else {
			// kapanış yoksa direkt text olarak ekle, baştaki <{ işareti eklenmiyor
			tokens = append(tokens, &Token{Type: TText, Value: part})
		}
	}

	return tokens
}
