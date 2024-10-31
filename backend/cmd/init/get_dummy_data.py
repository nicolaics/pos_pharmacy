import requests as rq


def get_users(BACKEND_ROOT, TOKEN):
    r = rq.get(
        f"http://{BACKEND_ROOT}/api/v1/user/all/all",
        headers={
            "Authorization": "Bearer " + TOKEN,
        },
    )
    print("users")
    print(r.json())


def get_patients(BACKEND_ROOT, TOKEN):
    r = rq.get(
        f"http://{BACKEND_ROOT}/api/v1/patient/all/all",
        headers={
            "Authorization": "Bearer " + TOKEN,
        },
    )
    print("patients")
    print(r.json())


def get_company_profile(BACKEND_ROOT, TOKEN):
    r = rq.get(
        f"http://{BACKEND_ROOT}/api/v1/company-profile",
        headers={
            "Authorization": "Bearer " + TOKEN,
        },
    )
    print("company profile")
    print(r.json())


def get_customers(BACKEND_ROOT, TOKEN):
    r = rq.get(
        f"http://{BACKEND_ROOT}/api/v1/customer/all",
        headers={
            "Authorization": "Bearer " + TOKEN,
        },
    )
    print("customers")
    print(r.json())


def get_doctors(BACKEND_ROOT, TOKEN):
    r = rq.get(
        f"http://{BACKEND_ROOT}/api/v1/doctor/all",
        headers={
            "Authorization": "Bearer " + TOKEN,
        },
    )
    print("doctors")
    print(r.json())


def get_suppliers(BACKEND_ROOT, TOKEN):
    r = rq.get(
        f"http://{BACKEND_ROOT}/api/v1/supplier/all/all",
        headers={
            "Authorization": "Bearer " + TOKEN,
        },
    )
    print("suppliers")
    print(r.json())


def get_medicines(BACKEND_ROOT, TOKEN):
    r = rq.get(
        f"http://{BACKEND_ROOT}/api/v1/medicine/all/all",
        headers={
            "Authorization": "Bearer " + TOKEN,
        },
    )
    print("medicines")
    print(r.json())


def get_invoices(BACKEND_ROOT, TOKEN):
    r = rq.post(
        f"http://{BACKEND_ROOT}/api/v1/invoice/all/all",
        json={"startDate": "2024-09-28 +0700GMT", "endDate": "2024-09-28 +0700GMT"},
        headers={
            "Authorization": "Bearer " + TOKEN,
        },
    )
    print("invoices")
    print(r.json())


def get_invoice_detail(BACKEND_ROOT, TOKEN):
    r = rq.post(
        f"http://{BACKEND_ROOT}/api/v1/invoice/detail",
        json={"id": 1},
        headers={
            "Authorization": "Bearer " + TOKEN,
        },
    )
    print("invoice detail id 1")
    print(r.json())


def get_po_invoices(BACKEND_ROOT, TOKEN):
    r = rq.post(
        f"http://{BACKEND_ROOT}/api/v1/invoice/purchase-order/all/all",
        json={"startDate": "2024-09-28 +0700GMT", "endDate": "2024-09-28 +0700GMT"},
        headers={
            "Authorization": "Bearer " + TOKEN,
        }
    )
    print("po invoices")
    print(r.json())


def get_po_invoice_detail(BACKEND_ROOT, TOKEN):
    r = rq.post(
        f"http://{BACKEND_ROOT}/api/v1/invoice/purchase-order/detail",
        json={"id": 1},
        headers={
            "Authorization": "Bearer " + TOKEN,
        }
    )
    print("po invoice detail id 1")
    print(r.json())


def get_prescriptions(BACKEND_ROOT, TOKEN):
    r = rq.post(
        f"http://{BACKEND_ROOT}/api/v1/prescription/all/all",
        json={
            "startDate": "2024-09-28 +0900GMT",
            "endDate": "2024-09-28 +0900GMT"
        },
        headers={
            "Authorization": "Bearer " + TOKEN,
        }
    )
    print("prescription")
    print(r.json())


def get_prescription_detail(BACKEND_ROOT, TOKEN):
    r = rq.post(
        f"http://{BACKEND_ROOT}/api/v1/prescription/detail",
        json={
            "id": 1
        },
        headers={
            "Authorization": "Bearer " + TOKEN,
        }
    )
    print("prescription detail id 1")
    print(r.json())


def get_productions(BACKEND_ROOT, TOKEN):
    r = rq.post(
        f"http://{BACKEND_ROOT}/api/v1/production/all/all",
        json={
            "startDate": "2024-09-28 +0900GMT",
            "endDate": "2024-09-28 +0900GMT"
        },
        headers={
            "Authorization": "Bearer " + TOKEN,
        }
    )
    print("productions")
    print(r.json())


def get_production_detail(BACKEND_ROOT, TOKEN):
    r = rq.post(
        f"http://{BACKEND_ROOT}/api/v1/production/detail",
        json={
            "id": 1
        },
        headers={
            "Authorization": "Bearer " + TOKEN,
        }
    )
    print("production detail id 1")
    print(r.json())


def get_purchase_invoices(BACKEND_ROOT, TOKEN):
    r = rq.post(
        f"http://{BACKEND_ROOT}/api/v1/invoice/purchase/all/all",
        json={
            "startDate": "2024-09-28 +0900GMT",
            "endDate": "2024-09-28 +0900GMT"
        },
        headers={
            "Authorization": "Bearer " + TOKEN,
        }
    )
    print("purchase invoices")
    print(r.json())


def get_purchase_invoice_detail(BACKEND_ROOT, TOKEN):
    r = rq.post(
        f"http://{BACKEND_ROOT}/api/v1/invoice/purchase/detail",
        json={
            "id": 3
        },
        headers={
            "Authorization": "Bearer " + TOKEN,
        }
    )
    print("purchase invoices id 3")
    print(r.json())


def main():
    # BACKEND_ROOT = input("enter backend host:port: ")
    BACKEND_ROOT = "localhost:9988"
    TOKEN = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhZG1pbiI6dHJ1ZSwiYXV0aG9yaXplZCI6dHJ1ZSwiZXhwaXJlZEF0IjoxNzI3NTI1MjE2LCJ0b2tlblV1aWQiOiIwMTkyMzg4NS1mN2JlLTdkZTMtYTBhYi01NDQ0YWRmYjhmZTAiLCJ1c2VySWQiOjJ9.LeORetCOQ0hs6Lw1eCHsbE-YZ2Kt5SjF2wKzSMKLBJw"
    get_users(BACKEND_ROOT, TOKEN)
    print()
    get_company_profile(BACKEND_ROOT, TOKEN)
    print()
    get_customers(BACKEND_ROOT, TOKEN)
    print()
    get_doctors(BACKEND_ROOT, TOKEN)
    print()
    get_patients(BACKEND_ROOT, TOKEN)
    print()
    get_suppliers(BACKEND_ROOT, TOKEN)
    print()
    get_medicines(BACKEND_ROOT, TOKEN)
    print()
    get_invoices(BACKEND_ROOT, TOKEN)
    print()
    get_invoice_detail(BACKEND_ROOT, TOKEN)
    print()
    get_po_invoices(BACKEND_ROOT, TOKEN)
    print()
    get_po_invoice_detail(BACKEND_ROOT, TOKEN)
    print()
    get_prescriptions(BACKEND_ROOT, TOKEN)
    print()
    get_prescription_detail(BACKEND_ROOT, TOKEN)
    print()
    get_productions(BACKEND_ROOT, TOKEN)
    print()
    get_production_detail(BACKEND_ROOT, TOKEN)
    print()
    get_purchase_invoices(BACKEND_ROOT, TOKEN)
    print()
    get_purchase_invoice_detail(BACKEND_ROOT, TOKEN)


if __name__ == "__main__":
    main()
