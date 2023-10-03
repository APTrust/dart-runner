package util

// NameValuePair contains a name and value. Use this
// to keep ordered lists of pairs, and lists that contain
// multiple values for the same named key. Go's maps
// can't do either of those things.
type NameValuePair struct {
	Name  string
	Value string
}

type NameValuePairList struct {
	Items []NameValuePair
}

func NewNameValuePairList() *NameValuePairList {
	return &NameValuePairList{
		Items: make([]NameValuePair, 0),
	}
}

func (list *NameValuePairList) Add(name, value string) {
	list.Items = append(list.Items, NameValuePair{name, value})
}

func (list *NameValuePairList) FirstMatching(name string) (NameValuePair, bool) {
	for _, nvp := range list.Items {
		if nvp.Name == name {
			return nvp, true
		}
	}
	return NameValuePair{}, false
}

func (list *NameValuePairList) AllMatching(name string) []NameValuePair {
	matches := make([]NameValuePair, 0)
	for _, nvp := range list.Items {
		if nvp.Name == name {
			matches = append(matches, nvp)
		}
	}
	return matches
}
