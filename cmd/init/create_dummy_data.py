import sys
import requests as rq
import random


def create_users(BACKEND_ROOT):
    r = rq.post(
        f"http://{BACKEND_ROOT}/api/v1/user/register",
        json={
            "name": "Adam adam",
            "password": "adampassword",
            "admin": True,
            "phoneNumber": "0819288176",
            "adminPassword": "I4geJeE0kSu5",
        },
    )
    r = rq.post(
        f"http://{BACKEND_ROOT}/api/v1/user/register",
        json={
            "name": "Bob bob",
            "password": "bobpassword",
            "admin": False,
            "phoneNumber": "0819288326",
            "adminPassword": "I4geJeE0kSu5",
        },
    )
    r = rq.post(
        f"http://{BACKEND_ROOT}/api/v1/user/register",
        json={
            "name": "Charlie charlie",
            "password": "charliepassword",
            "admin": False,
            "phoneNumber": "0813318326",
            "adminPassword": "I4geJeE0kSu5",
        },
    )
    r = rq.post(
        f"http://{BACKEND_ROOT}/api/v1/user/register",
        json={
            "name": "Delta delta",
            "password": "deltapassword",
            "admin": True,
            "phoneNumber": "081921296",
            "adminPassword": "I4geJeE0kSu5",
        },
    )
    # r = rq.post(f"http://{BACKEND_ROOT}/api/v1/user/register", json={
    #                     "name" : "Adam adam Two",
    #                     "email" : "AdamTwo@gmail.com",
    #                     "password" : "adamtwopassword"
    #                   })
    # r = rq.post(f"http://{BACKEND_ROOT}/api/v1/user/register", json={
    #                     "name" : "Bob bob Two",
    #                     "email" : "BobTwo@gmail.com",
    #                     "password" : "bobtwopassword"
    #                   })
    # r = rq.post(f"http://{BACKEND_ROOT}/api/v1/user/register", json={
    #                     "name" : "Charlie charlie Two",
    #                     "email" : "CharlieTwo@gmail.com",
    #                     "password" : "charlietwopassword"
    #                   })
    # r = rq.post(f"http://{BACKEND_ROOT}/api/v1/user/register", json={
    #                     "name" : "Delta delta Two",
    #                     "email" : "DeltaTwo@gmail.com",
    #                     "password" : "deltatwopassword"
    #                   })
    # r = rq.post(f"http://{BACKEND_ROOT}/api/v1/user/register", json={
    #                     "name" : "Zeta Zombo",
    #                     "email" : "ZetaZombo@gmail.com",
    #                     "password" : "zetazombopassword"
    #                   })
    # r = rq.post(f"http://{BACKEND_ROOT}/api/v1/user/register", json={
    #                     "name" : "Zeta Zombo Two",
    #                     "email" : "ZetaZomboTwo@gmail.com",
    #                     "password" : "zetazombotwopassword"
    #                   })


def create_company_profile(BACKEND_ROOT):
    r = rq.post(
        f"http://{BACKEND_ROOT}/api/v1/company-profile",
        json={
            "name": "Zulu",
            "address": "zulu abc",
            "businessNumber": "12345",
            "pharmacsit": "AA",
            "pharmacistLicenseNumber": "1239901",
        },
    )
    # for i in range(0, 15):
    #     for i in range(0, 10+1):
    #         r = rq.post(f"http://{BACKEND_ROOT}/api/v1/transaction/create", json={
    #                             "userid" : i,
    #                             "amount" : random.randint(1000, 1000000)
    #                           })


def create_customers(BACKEND_ROOT):
    r = rq.post(f"http://{BACKEND_ROOT}/api/v1/customer", json={"name": "Graph"})
    r = rq.post(f"http://{BACKEND_ROOT}/api/v1/customer", json={"name": "Alpha"})
    r = rq.post(f"http://{BACKEND_ROOT}/api/v1/customer", json={"name": "Beta"})
    r = rq.post(f"http://{BACKEND_ROOT}/api/v1/customer", json={"name": "Gama"})
    # print(r.json())


def create_doctors(BACKEND_ROOT):
    r = rq.post(f"http://{BACKEND_ROOT}/api/v1/doctor", json={"name": "Dr. Gray"})
    r = rq.post(f"http://{BACKEND_ROOT}/api/v1/doctor", json={"name": "Dr. Jay"})
    r = rq.post(f"http://{BACKEND_ROOT}/api/v1/doctor", json={"name": "Dr. Awesome"})
    r = rq.post(f"http://{BACKEND_ROOT}/api/v1/doctor", json={"name": "Dr. Pole"})
    r = rq.post(f"http://{BACKEND_ROOT}/api/v1/doctor", json={"name": "Dr. Ulala"})
    r = rq.post(f"http://{BACKEND_ROOT}/api/v1/doctor", json={"name": "Dr. Oscar"})


def create_patients(BACKEND_ROOT):
    r = rq.post(f"http://{BACKEND_ROOT}/api/v1/patient", json={"name": "Yankee"})
    r = rq.post(f"http://{BACKEND_ROOT}/api/v1/patient", json={"name": "Awesome"})
    r = rq.post(f"http://{BACKEND_ROOT}/api/v1/patient", json={"name": "Two"})


def create_medicines(BACKEND_ROOT):
    r = rq.post(f"http://{BACKEND_ROOT}/api/v1/medicine", json={
        "barcode": "1111",
        "name": "Ativan",
        "qty": 1000,
        "firstUnit": "BTL",
        "firstSubtotal": 11000,
        "firstDiscount": 1000,
        "firstPrice": 10000,
        "secondUnit": "BOX",
        "secondSubtotal": 112000,
        "secondPrice": 112000,
        "description": "test"
        })
    
    r = rq.post(f"http://{BACKEND_ROOT}/api/v1/medicine", json={
        "barcode": "2222",
        "name": "Panadol",
        "qty": 500,
        "firstUnit": "TAB",
        "firstSubtotal": 3320,
        "firstPrice": 3320,
        "description": "test"
        })
    
    r = rq.post(f"http://{BACKEND_ROOT}/api/v1/medicine", json={
        "barcode": "3333",
        "name": "Rhinos",
        "qty": 320,
        "firstUnit": "STP",
        "firstSubtotal": 19200,
        "firstDiscount": 200,
        "firstPrice": 19000
        })


def create_invoices(BACKEND_ROOT):
    r = rq.post(
        f"http://{BACKEND_ROOT}/api/v1/invoice",
        json={
            "number": 1,
            "customerId": 1,
            "subtotal": 10000,
            "discount": 100,
            "totalPrice": 9000,
            "paidAmount": 10000,
            "changeAmount": 1000,
            "paymentMethodName": "Cash",
            "invoiceDate": "2024-09-23",
            "medicineLists": [
                {
                    "medicineBarcode": "2222",
                    "medicineName": "Panadol",
                    "qty": 9,
                    "unit": "TAB",
                    "price": 3320,
                    "subtotal": 90000,
                }
            ],
        },
    )

    r = rq.post(
        f"http://{BACKEND_ROOT}/api/v1/invoice",
        json={
            "number": 2,
            "customerId": 5,
            "subtotal": 37000,
            "totalPrice": 37000,
            "paidAmount": 37000,
            "changeAmount": 0,
            "paymentMethodName": "Trasnfer",
            "invoiceDate": "2024-09-21",
            "medicineLists": [
                {
                    "medicineBarcode": "1111",
                    "medicineName": "Ativan",
                    "qty": 1,
                    "unit": "BTL",
                    "price": 11000,
                    "discount": 1000,
                    "subtotal": 30000,
                },
                {
                    "medicineBarcode": "1237738438",
                    "medicineName": "Rhinos",
                    "qty": 1,
                    "unit": "STP",
                    "price": 19000,
                    "discount": 3000,
                    "subtotal": 7000,
                }
            ],
        },
    )

def create_po_invoices(BACKEND_ROOT):
    r = rq.post(
        f"http://{BACKEND_ROOT}/api/v1/invoice/purchase-order",
        json={
            "number": 2,
            "companyId": 1,
            "supplierId": 3,
            "totalItems": 10,
            "invoiceDate": "2024-03-12",
            "purchaseOrderMedicineList": [
                {
                    "medicineBarcode": "1111",
                    "medicineName": "Ativan",
                    "orderQty": 7,
                    "receivedQty": 3,
                    "unit": "BTL",
                },
                {
                    "medicineBarcode": "3333",
                    "medicineName": "Rhinos",
                    "orderQty": 10,
                    "receivedQty": 10,
                    "unit": "TAB",
                    "remarks": "dfalks"
                }
            ],
        }
    )

def create_prescription(BACKEND_ROOT):
    r = rq.post(
        f"http://{BACKEND_ROOT}/api/v1/prescription",
        json={
            "invoice": {
                "number": 1,
                "userName": "admin",
                "customerName": "Alpha",
                "totalPrice": 30000,
                "invoiceDate": "2024-07-10"
            },
            "number": 100,
            "prescriptionDate": "2024-06-30",
            "patientName": "Yankee",
            "doctorName": "Dr. Play",
            "qty": 1,
            "price": 100000,
            "totalPrice": 100000,
            "prescriptionMedicineList": [
                {
                    "medicineBarcode": "1111",
                    "medicineName": "Ativan",
                    "qty": 7,
                    "unit": "BTL",
                    "price": 46200,
                    "subtotal": 46200
                },
                {
                    "medicineBarcode": "3333",
                    "medicineName": "Rhinos",
                    "qty": 10,
                    "unit": "TAB",
                    "price": 2200,
                    "discount": 200,
                    "subtotal": 2000
                }
            ]
        }
    )

    r = rq.post(
        f"http://{BACKEND_ROOT}/api/v1/prescription",
        json={
            "invoice": {
                "number": 1,
                "userName": "admin",
                "customerName": "Alpha",
                "totalPrice": 30000,
                "invoiceDate": "2024-07-10"
            },
            "number": 100,
            "prescriptionDate": "2024-06-30",
            "patientName": "Bpay",
            "doctorName": "Dr. Gray",
            "qty": 1,
            "price": 100000,
            "totalPrice": 100000,
            "prescriptionMedicineList": [
                {
                    "medicineBarcode": "1111",
                    "medicineName": "Ativan",
                    "qty": 7,
                    "unit": "TAB",
                    "price": 46200,
                    "subtotal": 46200
                },
                {
                    "medicineBarcode": "3333",
                    "medicineName": "Rhinos",
                    "qty": 10,
                    "unit": "TAB",
                    "price": 2200,
                    "discount": 200,
                    "subtotal": 2000
                }
            ]
        }
    )

def main():
    BACKEND_ROOT = input("enter backend host:port: ")
    create_users(BACKEND_ROOT)
    create_company_profile(BACKEND_ROOT)
    create_customers(BACKEND_ROOT)
    create_patients(BACKEND_ROOT)
    create_medicines(BACKEND_ROOT)
    create_invoices(BACKEND_ROOT)


if __name__ == "__main__":
    main()
