// src/pages/HomePage.jsx
import { useNavigate } from "react-router-dom";

export default function HomePage() {
    const navigate = useNavigate();

    return (
        <div
            style={{
                height: "100vh",
                display: "flex",
                flexDirection: "column",
                justifyContent: "center",
                alignItems: "center",
                textAlign: "center",
            }}
        >
            <h1 style={{ fontSize: "2rem", marginBottom: "3rem" }}>
                Welcome to <span style={{ color: "#22c55e" }}>ReceiptVault 🧾</span>
            </h1>

            <div
                style={{
                    display: "flex",
                    gap: "2rem",
                    justifyContent: "center",
                    width: "50%",
                }}
            >
                <button
                    onClick={() => navigate("/receipts")}
                    style={{
                        flex: 1,
                        padding: "1rem",
                        backgroundColor: "#3b82f6",
                        color: "white",
                        border: "none",
                        borderRadius: "0.5rem",
                        fontWeight: "bold",
                        cursor: "pointer",
                    }}
                >
                    Get All Receipts
                </button>

                <button
                    onClick={() => navigate("/upload")}
                    style={{
                        flex: 1,
                        padding: "1rem",
                        backgroundColor: "#22c55e",
                        color: "white",
                        border: "none",
                        borderRadius: "0.5rem",
                        fontWeight: "bold",
                        cursor: "pointer",
                    }}
                >
                    Upload Receipts
                </button>
            </div>
        </div>
    );
}
