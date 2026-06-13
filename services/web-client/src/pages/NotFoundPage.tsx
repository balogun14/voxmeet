import { useNavigate } from "react-router-dom";

export default function NotFoundPage() {
  const navigate = useNavigate();
  return (
    <div className="min-h-screen flex flex-col items-center justify-center gap-4">
      <h1 className="text-4xl font-bold">404</h1>
      <p className="text-neutral-400">Page not found</p>
      <button
        className="text-blue-400 underline"
        onClick={() => navigate("/")}
      >
        Go home
      </button>
    </div>
  );
}
