import React from 'react';
import { Review } from '../../api/types';
import { ReviewItem } from './ReviewItem';

type Props = {
    reviews: Review[];
};

export const ReviewList: React.FC<Props> = ({ reviews }) => {
    if (!reviews.length) {
        return <p>Пока нет отзывов.</p>;
    }

    return (
        <div className="review-list">
            {reviews.map((r) => (
                <ReviewItem key={r.id} review={r} />
            ))}
        </div>
    );
};
