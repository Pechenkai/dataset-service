import React from 'react';
import { Review } from '../../api/types';
import { RatingChip } from '../ui/RatingChip';

type Props = {
    review: Review;
    username?: string; // если пока нет user lookup — прокинешь "User #id"
};

export const ReviewItem: React.FC<Props> = ({ review, username }) => {
    const name = username ?? `User #${review.user_id}`;

    return (
        <div className="review-item">
            <div className="review-item__head">
                <div className="review-item__user">{name}</div>
                <RatingChip label={`Rating ${review.rating}`} />
            </div>

            <div className="review-item__body">
                {review.text || 'Без комментариев'}
            </div>
        </div>
    );
};
