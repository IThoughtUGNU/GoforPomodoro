package main

type AlreadySubscribed struct{}

func (_ AlreadySubscribed) Error() string {
	return "you were already subscribed"
}

type AlreadyUnsubscribed struct{}

func (_ AlreadyUnsubscribed) Error() string {
	return "you were already unsubscribed"
}

type SubscriptionError struct{}

func (_ SubscriptionError) Error() string {
	return "cannot subscribe"
}

type OperationError struct{}

func (_ OperationError) Error() string {
	return "error with this operation right now"
}
