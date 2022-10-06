// This file is part of GoforPomodoro.
//
// GoforPomodoro is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// GoforPomodoro is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with GoforPomodoro.  If not, see <http://www.gnu.org/licenses/>.

package domain

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
