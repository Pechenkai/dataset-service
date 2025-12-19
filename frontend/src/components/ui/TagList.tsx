import React from 'react';
import { Tag } from './Tag';

type Props = {
    tags: string[];
    className?: string;
};

export const TagList: React.FC<Props> = ({ tags, className = '' }) => {
    if (!tags.length) return null;

    return (
        <div className={['ui-tag-list', className].filter(Boolean).join(' ')}>
            {tags.map((t) => (
                <Tag key={t}>{t}</Tag>
            ))}
        </div>
    );
};
