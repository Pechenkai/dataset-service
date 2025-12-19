import React, { useMemo, useState } from 'react';
import { Button } from '../ui/Button';
import { Input } from '../ui/Input';
import { TextArea } from '../ui/TextArea';

type Props = {
    onSubmit: (payload: { rating: number; text?: string }) => Promise<void> | void;
    isSubmitting?: boolean;
    initialRating?: number;
    initialText?: string;
};

export const ReviewForm: React.FC<Props> = ({
                                                onSubmit,
                                                isSubmitting = false,
                                                initialRating = 5,
                                                initialText = ''
                                            }) => {
    const [ratingRaw, setRatingRaw] = useState(String(initialRating));
    const [text, setText] = useState(initialText);
    const [error, setError] = useState<string>('');

    const rating = useMemo(() => {
        const n = Number(ratingRaw);
        if (!Number.isFinite(n)) return 0;
        return Math.max(1, Math.min(5, Math.trunc(n)));
    }, [ratingRaw]);

    const canSubmit = rating >= 1 && rating <= 5 && !isSubmitting;

    const submit = async (e: React.FormEvent) => {
        e.preventDefault();
        setError('');

        if (rating < 1 || rating > 5) {
            setError('Rating должен быть в диапазоне 1–5.');
            return;
        }

        try {
            await onSubmit({ rating, text: text.trim() ? text.trim() : undefined });
            setText('');
            setRatingRaw(String(initialRating));
        } catch {
            setError('Не удалось отправить отзыв.');
        }
    };

    return (
        <form className="review-form" onSubmit={submit}>
            <div className="review-form__grid">
                <Input
                    label="Rating"
                    type="number"
                    min={1}
                    max={5}
                    value={ratingRaw}
                    onChange={(e) => setRatingRaw(e.target.value)}
                    required
                />
                <TextArea
                    label="Comment"
                    placeholder="Enter dataset description..."
                    value={text}
                    onChange={(e) => setText(e.target.value)}
                    rows={4}
                />
            </div>

            <div className="review-form__actions">
                <Button type="submit" variant="primary" disabled={!canSubmit}>
                    Publish
                </Button>
                {error && <div className="review-form__error">{error}</div>}
            </div>
        </form>
    );
};
