import { useState } from "react";
import axios from "axios";

export default function UploadFlow() {
    const [step, setStep] = useState(1);
    const [file, setFile] = useState(null);
    const [uploadInfo, setUploadInfo] = useState(null); // holds file_id & file_name
    const [receiptData, setReceiptData] = useState(null);

    const handleFileUpload = async (e) => {
        e.preventDefault();
        if (!file) return alert("Please select a file");

        const formData = new FormData();
        formData.append("file", file);

        try {
            const res = await axios.post("http://localhost:8000/upload", formData);
            setUploadInfo({
                file_id: res.data.file_id,
                file_name: res.data.file_name,
            });
            setStep(2);
        } catch (err) {
            console.error("Upload failed:", err);
            alert("Failed to upload file");
        }
    };

    return (
        <div style={{ maxWidth: "600px", margin: "2rem auto", padding: "1rem" }}>
            <h2 style={{ textAlign: "center", marginBottom: "2rem" }}>📤 Upload Receipt Workflow</h2>

            {/* Step 1 */}
            <div style={sectionStyle}>
                <h3>Step 1: Upload Receipt File</h3>
                <form onSubmit={handleFileUpload}>
                    <input
                        type="file"
                        accept="application/pdf"
                        onChange={(e) => setFile(e.target.files[0])}
                        required
                    />
                    <br />
                    <button type="submit" style={{ ...btnStyle, marginTop: "1rem" }}>
                        Upload
                    </button>
                </form>
            </div>

            {/* Step 2 */}
            {step >= 2 && uploadInfo && (
                <div style={sectionStyle}>
                    <h3>Step 2: Validate Receipt</h3>
                    <button
                        style={btnStyle}
                        onClick={async () => {
                            try {
                                const res = await axios.post("http://localhost:8000/validate", {
                                    file_id: uploadInfo.file_id,
                                    file_name: uploadInfo.file_name,
                                });

                                if (res.data.is_valid) {
                                    alert("✅ File is valid!");
                                    setStep(3);
                                } else {
                                    alert("❌ File is invalid. Reason: " + (res.data.invalid_reason || "Unknown"));
                                }
                            } catch (err) {
                                console.error(err);
                                alert("⚠️ Error while validating file.");
                            }
                        }}
                    >
                        Validate
                    </button>
                </div>
            )}

            {/* Step 3 */}
            {step >= 3 && (
                <div style={sectionStyle}>
                    <h3>Step 3: Process Receipt</h3>
                    <button
                        style={btnStyleGreen}
                        onClick={async () => {
                            try {
                                const res = await axios.post("http://localhost:8000/process", {
                                    file_id: uploadInfo.file_id,
                                    file_name: uploadInfo.file_name,
                                });
                                setReceiptData(res.data);
                                setStep(4);
                            } catch (err) {
                                alert("Processing failed");
                            }
                        }}
                    >
                        Process
                    </button>
                </div>
            )}

            {/* Step 4 - Result */}
            {step === 4 && receiptData && (
                <div style={{ ...sectionStyle, backgroundColor: "#f0f9f0" }}>
                    <h3>✅ Receipt Processed</h3>
                    <p style={{ fontWeight: "bold", color: "green" }}>{receiptData.message}</p>
                    <div style={resultBoxStyle}>
                        <p><strong>Merchant Name:</strong> {receiptData.merchant_name}</p>
                        <p><strong>Purchased At:</strong> {new Date(receiptData.purchased_at).toLocaleString()}</p>
                        <p><strong>Total Amount:</strong> ₹{receiptData.total_amount}</p>
                    </div>
                </div>
            )}
        </div>
    );
}

const sectionStyle = {
    border: "1px solid #ccc",
    padding: "1.5rem",
    borderRadius: "8px",
    marginBottom: "1.5rem",
    backgroundColor: "#f9f9f9"
};

const btnStyle = {
    backgroundColor: "#007bff",
    color: "#fff",
    border: "none",
    padding: "0.5rem 1.5rem",
    borderRadius: "4px",
    cursor: "pointer",
};

const btnStyleGreen = {
    ...btnStyle,
    backgroundColor: "#28a745"
};

const resultBoxStyle = {
    border: "1px dashed #ccc",
    padding: "1rem",
    marginTop: "1rem",
    borderRadius: "6px",
    backgroundColor: "#fff"
};
