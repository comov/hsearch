from django.core.paginator import Paginator
from django.http import JsonResponse

from hsearch.models import Apartment


def apartment_list(request):
    try:
        page = int(request.GET.get("page", 1))
    except ValueError:
        return JsonResponse({"message": "The 'page' parameter must be an integer."}, status=400)

    try:
        pear_page = int(request.GET.get("pear_page", 100))
    except ValueError:
        return JsonResponse({"message": "The 'pear_page' parameter must be an integer."}, status=400)

    available_fields = Apartment.available_fields
    available_relations_fields = Apartment.available_relations_fields

    order = str(request.GET.get("order", "id"))
    ordering = order in available_fields
    reverse_ordering = order.startswith("-") and order[1:] in available_fields
    if not ordering and not reverse_ordering:
        return JsonResponse({
            "message": f"Sorting can only be used on the following fields: {available_fields}",
        }, status=400)

    _fields = request.GET.get("fields", "")
    fields_list = (_fields and str(_fields).split(",")) or available_fields + available_relations_fields
    obj_fields = list(set(fields_list) & set(available_fields))
    relations_fields, relations = [], set()
    for relation in set(fields_list) & set(available_relations_fields):  # type: str
        if "__" not in relation:
            continue
        model, _field = relation.split("__")[:2]
        relations.add(model)
        relations_fields.append(_field)

    relations = list(relations)

    queryset = Apartment.objects.prefetch_related(*relations).only(*obj_fields).order_by(order)

    paginator = Paginator(queryset, per_page=pear_page)
    return JsonResponse({
        "results": [
            i.to_dict(obj_fields, relations, relations_fields)
            for i in paginator.get_page(page)
        ] if paginator.num_pages >= page > 0 else [],
        "total": paginator.count,
        "pear_page": pear_page,
        "total_page": paginator.num_pages,
        "current_page": page,
        "order": order,
    })
