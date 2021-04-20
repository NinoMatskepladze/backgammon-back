package structs

type card struct {
	Index int
	Suit  int // ყვავი, წკენტა და ასე შ.
	Kind  int // K A 10 ...
}

func GetCard(index int) card {
	var card card
	card.Index = index
	card.Suit = index / 13
	card.Kind = index % 13
	return card
}

type buraCard struct {
	Index int
	Suit  int // ყვავი, წკენტა და ასე შ.
	Kind  int // K A 10 ...
	Power int
	Score int
}

func GetBuraCard(index int) buraCard {
	var card buraCard
	card.Index = index
	card.Suit = index / 13
	card.Kind = index % 13

	if card.Kind == 8 {
		card.Power = 11 // თუკი 10 იანია დააყენოს ტუზის წინ ძალით
	} else if card.Kind > 8 && card.Kind < 12 {
		card.Power = card.Kind - 1 // J Q K დააყენოს 10 იანზე დაბლა
	} else {
		card.Power = card.Kind // სხვა ყველა შემთხვევაში მისცეს კარტს ის ძალა რა კარტიცაა.
	}

	if card.Kind < 8 {
		card.Score = 0 // 10 იანამდე ყველას 0 ქულა აქვს.
	} else if card.Kind == 12 {
		card.Score = 11 // ტუზს 11 ქულა აქვს
	} else if card.Kind == 8 {
		card.Score = 10 // 10 იანს 10
	} else if card.Kind > 8 && card.Kind < 12 {
		card.Score = card.Power - 6 // დანარჩენებს 6 ით ნაკლები რა ძალაც აქვთ ანუ 2/3/4
	}
	return card
}
