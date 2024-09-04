import "@/index.css";
import * as React from "react";
import { APIErrorBody } from "@/lib/types";

function App() {
  const [response, setResponse] = React.useState(null);
  const [error, setError] = React.useState<string | APIErrorBody | null>(null);

  const pingAPI = async () => {
    try {
      const res = await fetch(import.meta.env.VITE_API_URL);
      if (!res.ok) {
        // Try to parse the error body as an APIErrorBody
        const errorBody: APIErrorBody = await res.json();
        throw errorBody;
      }
      const data = await res.json();
      setResponse(data);
    } catch (err: unknown) {
      if (err instanceof Error) {
        setError(err.message);
      } else if (
        typeof err === "object" &&
        err !== null &&
        "status_code" in err
      ) {
        setError(err as APIErrorBody);
      } else {
        setError("An unknown error occurred");
      }
    }
  };

  return (
    <>
      <div>
        <button onClick={pingAPI}>Ping API</button>
        {response && <div>Response: {JSON.stringify(response)}</div>}
        {error && (
          <div>
            Error: {typeof error === "string" ? error : JSON.stringify(error)}
          </div>
        )}
      </div>
    </>
  );
}

export default App;
