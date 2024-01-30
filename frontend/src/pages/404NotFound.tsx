import React from 'react';

const NotFoundPage: React.FC = () => {
    return (
        <div className="flex flex-col items-center justify-center h-screen bg-yellow-400 text-black">
            <h1 className="text-4xl font-bold mb-4">404 - Page Not Found</h1>
            <p className="text-lg">Oops! The page you are looking for does not exist.</p>
        </div>
    );
};

export default NotFoundPage;