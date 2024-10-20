// history.tsx
import React, { useState } from "react";
import { FaPills } from "react-icons/fa";
import { useNavigate } from "react-router-dom";
import "./History.css";

const HistoryPage: React.FC = () => {
  const navigate = useNavigate();
  const [startDate, setStartDate] = useState("");
  const [endDate, setEndDate] = useState("");
  const [historyData, setHistoryData] = useState<any[]>([]); // Array to hold fetched history data

  const handleStartDateChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setStartDate(e.target.value);
  };

  const handleEndDateChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setEndDate(e.target.value);
  };

  const fetchHistory = () => {
    const token = sessionStorage.getItem("token");
    const url = `http://BACKEND_URL/history?start=${startDate}&end=${endDate}`; // Replace with backend URL

    fetch(url, {
      method: "GET",
      headers: {
        "Content-Type": "application/json",
        Authorization: `Bearer ${token}`,
      },
    })
      .then((response) => {
        if (!response.ok) {
          throw new Error("Error fetching history data");
        }
        return response.json();
      })
      .then((data) => {
        setHistoryData(data);
      })
      .catch((error) => {
        console.error("Error:", error);
      });
  };

  const returnInventory = () => {
    navigate("/inventory");
  };

  return (
    <div className="history-page">
      <h1>Inventory History</h1>
      <div className="date-range-container">
        <input type="date" value={startDate} onChange={handleStartDateChange} />
        <input type="date" value={endDate} onChange={handleEndDateChange} />
        <button onClick={fetchHistory} className="history-button">
          Filter
        </button>
      </div>

      <div className="history-table-container">
        <table id="history-data-table" border={1}>
          <thead>
            <tr>
              <th className="history-id-column">ID</th>
              <th className="history-date-column">Date</th>
              <th className="history-in-column">In Quantity</th>
              <th className="history-unit-column">Unit(s)</th>
              <th className="history-out-column">Out Quantity</th>
              <th className="history-unit-column">Unit(s)</th>
              <th className="history-cashier-column">Cashier</th>
              <th className="history-supplier-column">Supplier</th>
            </tr>
          </thead>
          <tbody>
            {historyData.map((item) => (
              <tr key={item.id}>
                <td>{item.id}</td>
                <td>{item.date}</td>
                <td>{item.barcode}</td>
                <td>{item.name}</td>
                <td>{item.action}</td>
              </tr>
            ))}
          </tbody>
        </table>

        <div className="inventory-return-btns-grp">
          <button
            onClick={returnInventory}
            className="view-inventory-home-button"
          >
            <FaPills size={30} id="view-inventory-home-icon" />
            Back to Inventory
          </button>
        </div>
      </div>
    </div>
  );
};

export default HistoryPage;
