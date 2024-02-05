import { useState } from 'react';

const useRequest = () => {
  const [data, setData] = useState(null);
  const [error, setError] = useState(null);
  const [isLoading, setIsLoading] = useState(false);

  const performRequest = async (endpoint: string, method: string, requestData?: object) => {
    try {
      setIsLoading(true);

      const requestOptions = {
        method,
        headers: {
          'Content-Type': 'application/json',
          // Add any additional headers if needed
        },
        body: method !== 'GET' ? JSON.stringify(requestData) : null,
      };

      const response = await fetch(endpoint, requestOptions);

      if (!response.ok) {
        throw new Error(`Request failed with status ${response.status}`);
      }

      const responseData = await response.json();
      setData(responseData);
      setError(null);
    } catch (err) {
      setData(null);
      setError(err.message);
    } finally {
      setIsLoading(false);
    }
  };

  return {
    data,
    error,
    isLoading,
    performRequest,
  };
};

export default useRequest;

