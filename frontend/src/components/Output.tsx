const Output = ({data, error, isLoading, invalidMethod}) => {
	
	if (invalidMethod) {
		console.error(invalidMethod)
	}

  return (
    <div className="bg-black p-4 m-5 w-1/3 h-full rounded-lg shadow-lg text-white">
      <pre className="text-sm font-mono whitespace-pre-wrap overflow-x-auto">
        Output:
      </pre>
			<pre className="text-sm font-mono whitespace-pre-wrap overflow-x-auto">
				{isLoading}
        { !isLoading &&
				error ? error : data}
      </pre>
    </div>
  );
};

export default Output;